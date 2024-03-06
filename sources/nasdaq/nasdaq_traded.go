// sources/nasdaq/nasdaq_traded.go
// Copyright (c) 2024 Neomantra BV
//
// Specs:
//   http://www.nasdaqtrader.com/trader.aspx?id=symboldirdefs
//   https://www.nasdaqtrader.com/Trader.aspx?id=CQSSymbolConvention
//
// Source:
//   ftp://ftp.nasdaqtrader.com/symboldirectory/nasdaqtraded.txt
//
// Sample:
// Nasdaq Traded|Symbol|Security Name|Listing Exchange|Market Category|ETF|Round Lot Size|Test Issue|Financial Status|CQS Symbol|NASDAQ Symbol|NextShares
// Y|A|Agilent Technologies, Inc. Common Stock|N| |N|100|N||A|A|N
//

package nasdaq

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/NimbleMarkets/ticker_autocomplete/internal/nimble"
	"github.com/jlaffaye/ftp"
	"github.com/jszwec/csvutil" // instead of gocsv because csvutil supports empty headers
)

const NASDAQ_TRADED_URL = "ftp://ftp.nasdaqtrader.com/symboldirectory/nasdaqtraded.txt"
const NASDAQ_FTP_HOSTNAME = "ftp.nasdaqtrader.com"
const NASDAQ_FTP_TRADED_PATH = "symboldirectory/nasdaqtraded.txt"
const NASDAQ_TRADED_FILENAME = "nasdaqtraded.txt"

///////////////////////////////////////////////////////////////////////////////

// NasdaqTraded is a CSV row of the Nasdaq Traded file
// Nasdaq Traded|Symbol|Security Name|Listing Exchange|Market Category|ETF|Round Lot Size|Test Issue|Financial Status|CQS Symbol|NASDAQ Symbol|NextShares
type NasdaqTraded struct {
	Traded          string `csv:"Nasdaq Traded" db:"nasdaq_traded" json:"nasdaq_traded"`
	Symbol          string `csv:"Symbol" db:"symbol" json:"symbol"`
	Name            string `csv:"Security Name" db:"security_name" json:"security_name"`
	ListingExchange string `csv:"Listing Exchange" db:"listing_exchange" json:"listing_exchange"`
	Category        string `csv:"Market Category" db:"market_category" json:"market_category"`
	IsETF           string `csv:"ETF" db:"etf" json:"etf"`
	RoundLotSize    int    `csv:"Round Lot Size" db:"round_lot_size" json:"round_lot_size"`
	IsTestIssue     string `csv:"Test Issue" db:"test_issue" json:"test_issue"`
	FinancialStatus string `csv:"Financial Status" db:"financial_status" json:"financial_status"`
	CqsSymbol       string `csv:"CQS Symbol" db:"cqs_symbol" json:"cqs_symbol"`
	NasdaqSymbol    string `csv:"NASDAQ Symbol" db:"nasdaq_symbol" json:"nasdaq_symbol"`
	IsNextShare     string `csv:"NextShares" db:"nextshares" json:"nextshares"`
	Type            string `csv:"-" db:"security_type" json:"security_type"` // not in CSV
	Tape            string `csv:"-" db:"tape" json:"tape"`                   // not in CSV
}

///////////////////////////////////////////////////////////////////////////////

// Returns Tape "A", "B", or "C" for the given listing exchange code.
// Returns "" for unknown
func GetTapeForListingExchange(listingExchange string) string {
	switch listingExchange {
	case "N": // New York Stock Exchange (NYSE)
		return "A"
	case "Q": // NASDAQ
		return "C"
	case "A": // NYSE MKT
		return "B"
	case "P": // NYSE ARCA
		return "B"
	case "Z": // BATS Global Markets (BATS)
		return "B"
	case "V": // Investors' Exchange, LLC (IEXG)
		return "B"
	default:
		return ""
	}
}

///////////////////////////////////////////////////////////////////////////////

func FetchNasdaqTraded() ([]NasdaqTraded, error) {
	// Check the cache
	nqt_bytes, cacheFilename, err := checkCache()
	if err != nil {
		// Uncached, so download file from FTP
		nqt_bytes, err = FtpDownloadNasdaqTradedFile()
		if err != nil {
			return nil, err
		}

		// Save to cache... if we fail here, it's not an error
		os.WriteFile(cacheFilename, nqt_bytes, 0644)
	}

	// Parse the raw data into NasdaqTraded structs
	nqts, err := ParseNasdaqTradedBytes(nqt_bytes)
	if err != nil {
		return nil, err
	}

	return nqts, nil
}

///////////////////////////////////////////////////////////////////////////////

func checkCache() ([]byte, string, error) {
	nimbleDir, err := nimble.GetNimbleDir()
	if err != nil {
		return nil, "", fmt.Errorf("unable to find Nimble directory: %v", err)
	}

	// TODO: check age

	cachedFilename := filepath.Join(nimbleDir, NASDAQ_TRADED_FILENAME)
	nqt_bytes, err := os.ReadFile(cachedFilename)
	return nqt_bytes, cachedFilename, err
}

///////////////////////////////////////////////////////////////////////////////

// Given a nasdaqtraded.txt input file bytes, returns an array of the parsed data.
// Returns nil and error upon any failure.
func ParseNasdaqTradedBytes(nqt_bytes []byte) ([]NasdaqTraded, error) {
	// We need to remove the last line:
	//   File Creation Time: 0306202412:12|||||
	// This messes up parsing due to incorrect number of fields
	lastIndex := bytes.LastIndex(nqt_bytes, []byte("File Creation Time"))
	if lastIndex > 0 {
		nqt_bytes = nqt_bytes[:lastIndex-1]
	}

	// Setup CSV parsing
	bytesReader := bytes.NewReader(nqt_bytes)
	csvReader := csv.NewReader(bytesReader)
	csvReader.Comma = '|' // NASDAQ files are pipe | delimited
	csvReader.FieldsPerRecord = 12

	decoder, err := csvutil.NewDecoder(csvReader)
	if err != nil {
		return nil, fmt.Errorf("nasdaqtraded CSV decoder create failed: %v", err)
	}

	// sometimes IsEtf field has ' ' instead of 'N'
	decoder.Map = func(field, column string, v interface{}) string {
		if column == "ETF" && field == " " {
			return "N"
		}
		return field
	}

	var nasdaqTraded []NasdaqTraded
	for {
		var row NasdaqTraded
		if err := decoder.Decode(&row); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("nasdaqtraded CSV file parse failed: %v", err)
		}
		row.Tape = GetTapeForListingExchange(row.ListingExchange)
		// TODO: row.Type
		nasdaqTraded = append(nasdaqTraded, row)
	}
	return nasdaqTraded, nil
}

///////////////////////////////////////////////////////////////////////////////

// Downloads the nasdaqtraded.txt file anonymously from NASDAQ's FTP.
// Returns the downloaded bytes or an error on failure.
func FtpDownloadNasdaqTradedFile() ([]byte, error) {
	// FTP Read
	ftpClient, err := ftp.Dial(
		fmt.Sprintf("%s:%d", NASDAQ_FTP_HOSTNAME, 21),
		ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return nil, err
	}
	defer ftpClient.Quit()

	err = ftpClient.Login("anonymous", "anonymous")
	if err != nil {
		return nil, err
	}

	resp, err := ftpClient.Retr(NASDAQ_FTP_TRADED_PATH)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	buf, err := io.ReadAll(resp)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
