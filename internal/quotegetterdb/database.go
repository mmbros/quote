package quotegetterdb

import (
	"bytes"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

// QuoteDatabase handles the database that store and retrieve quote informations.
type QuoteDatabase struct {
	dns string
	db  *sql.DB
}

// QuoteRecord is the record stored in the quote database.
type QuoteRecord struct {
	ID        int
	Isin      string
	Source    string
	Timestamp time.Time
	Date      time.Time
	Price     float32
	Currency  string
	URL       string
	ErrMsg    string
}

func (qr *QuoteRecord) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("{id=%d, isin=%q, source=%q", qr.ID, qr.Isin, qr.Source))
	buf.WriteString(fmt.Sprintf(", timestamp=%s", qr.Timestamp.UTC()))

	if !qr.Date.IsZero() {
		buf.WriteString(fmt.Sprintf(", date=%s", qr.Date.Format("2006-01-02")))
	}
	if len(qr.Currency) > 0 {
		buf.WriteString(fmt.Sprintf(", price=%.3f %s", qr.Price, qr.Currency))
	}
	if len(qr.URL) > 0 {
		buf.WriteString(fmt.Sprintf(", url=%q", qr.URL))
	}
	if len(qr.ErrMsg) > 0 {
		buf.WriteString(fmt.Sprintf(", err=%q", qr.ErrMsg))
	}
	buf.WriteString("}")
	return buf.String()
}

type errorQuoteDatabase struct {
	msg string
	err error
}

func newError(msg string, err error) error {
	return &errorQuoteDatabase{msg, err}
}

func (e *errorQuoteDatabase) Error() string {
	return fmt.Sprintf("%s: %s", e.msg, e.err)
}

func (e *errorQuoteDatabase) Unwrap() error {
	return e.err
}

func (e *errorQuoteDatabase) Msg() string {
	return e.msg
}

// ToNullString invalidates a sql.NullString if empty, validates if not empty
func ToNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

// ToNullTime invalidates a sql.NullTime if IsZero, validates otherwise
func ToNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{
		Time:  t,
		Valid: !t.IsZero(),
	}
}

// ToNullFloat64 invalidates a sql.NullFloat64 if 0, validates otherwise
func ToNullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{
		Float64: f,
		Valid:   f != 0.0,
	}
}

/*
func extractDir(dns string) string {
	pathname := dns

	if strings.Contains(pathname, ":memory:") || strings.Contains(pathname, "mode=memory") {
		return ""
	}
	if strings.HasPrefix(pathname, "file:") {
		pathname = pathname[5:]
	}
	if i := strings.IndexRune(pathname, '?'); i >= 0 {
		pathname = pathname[:i]
	}

	return path.Dir(pathname)
}
*/

// Open the quote database. Try to create the database if not exists.
func Open(dns string) (*QuoteDatabase, error) {
	/*
		folder := extractDir(dns)
		if folder != "" {
			if _, err := os.Stat(folder); os.IsNotExist(err) {
				os.MkdirAll(folder, os.ModeDir+0700)
			}
		}
	*/
	db, err := sql.Open("sqlite3", dns)
	if err == nil {
		err = db.Ping()
	}
	if err != nil {
		return nil, newError(fmt.Sprintf("Opening quotes database %q", dns), err)
	}

	qdb := &QuoteDatabase{dns, db}

	err = qdb.initDatabase()
	if err != nil {
		qdb.Close()
		return nil, err
	}

	return qdb, nil
}

// Close the quote database
func (qdb *QuoteDatabase) Close() error {
	if qdb == nil || qdb.db == nil {
		return nil
	}

	if err := qdb.db.Close(); err != nil {
		return err
	}

	return nil
}

func (qdb *QuoteDatabase) initDatabase() error {
	if e := qdb.createTableQuotes(); e != nil {
		return e
	}
	// if e := qdb.createViewQuotes(); e != nil {
	// 	return e
	// }
	return nil
}

func (qdb *QuoteDatabase) createTableQuotes() error {
	/*
		crea un unique index sui campi (isin, source, datestamp, date)
		- datestamp e' il timestamp con la sola data, senza orario
		- date e' reso not null per evitare di far fallire il controllo di unique

		in caso di insert con gli stessi valori di (isin, source, datestamp, date)
		il nuovo record sostituisce il vecchio mediante la clausola
		   INSERT OR REPLACE INTO quotes

		In questo modo, a parita' di isin, source e datastamp,
		sara' presente un solo record per ogni data
		Ad esempio sara' possibile avere:
		  ISIN          SOURCE     DATASTAMP   DATE
		  isin00001234  source.it  2020-10-01  2020-09-30
		  isin00001234  source.it  2020-10-01  2020-09-29
		  isin00001234  source.it  2020-10-01  0001-01-01  (zero date)
		ma non
		  ISIN          SOURCE     DATASTAMP   DATE
		  isin00001234  source.it  2020-10-01  2020-09-30
		  isin00001234  source.it  2020-10-01  2020-09-30
		e non
		  ISIN          SOURCE     DATASTAMP   DATE
		  isin00001234  source.it  2020-10-01  0001-01-01  (zero date)
		  isin00001234  source.it  2020-10-01  0001-01-01  (zero date)
	*/

	// create table if not exists
	sql := `CREATE TABLE IF NOT EXISTS quotes(
id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
isin TEXT NOT NULL,
source TEXT NOT NULL,
datestamp DATETIME NOT NULL,
timestamp DATETIME NOT NULL,
date DATE NOT NULL,
price DOUBLE,
currency TEXT,
url TEXT,
errmsg TEXT
);
`

	_, err := qdb.db.Exec(sql)
	if err != nil {
		return newError("Create table 'quotes'", err)
	}

	// create index if not exists
	sql = `CREATE UNIQUE INDEX IF NOT EXISTS idx_quotes_isin_source_dates 
ON quotes (isin, source, datestamp, date);`
	_, err = qdb.db.Exec(sql)
	if err != nil {
		return newError("Create index 'idx_quotes_isin_source_dates'", err)
	}

	return nil
}

// func (qdb *QuoteDatabase) createViewQuotes() error {

// 	// create table if not exists
// 	sql := `CREATE VIEW IF NOT EXISTS v_quotes
// AS
// select *,
// case when (date is null OR price is null OR currency is null) then 0 else 1 end AS status
// from quotes
// ;
// `

// 	_, err := qdb.db.Exec(sql)
// 	if err != nil {
// 		return newError("Create view 'v_quotes'", err)
// 	}

// 	return nil
// }

// InsertQuotes insert the quotes in the quotes database.
func (qdb *QuoteDatabase) InsertQuotes(items ...*QuoteRecord) error {
	sql := `INSERT OR REPLACE INTO quotes(
datestamp,
timestamp,
isin,
source,
date,
price,
currency,
url,
errmsg
) values(?, ?, ?, ?, ?, ?, ?, ?, ?)
`
	stmt, err := qdb.db.Prepare(sql)
	if err != nil {
		return newError("Insert quote", err)
	}
	defer stmt.Close()

	for _, i := range items {
		timestamp := i.Timestamp
		if timestamp.IsZero() {
			timestamp = time.Now()
		}
		// set datestamp
		year, month, day := timestamp.Date()
		datestamp := time.Date(year, month, day, 0, 0, 0, 0, timestamp.Location())

		_, err = stmt.Exec(datestamp, timestamp, i.Isin, i.Source,
			i.Date, // ToNullTime(i.date),
			ToNullFloat64(float64(i.Price)),
			ToNullString(i.Currency),
			ToNullString(i.URL),
			ToNullString(i.ErrMsg))
		if err != nil {
			return newError("Insert quote", err)
		}
	}

	return nil
}

/*
// SelectAllQuotes select all the quotes of the database.
func (qdb *QuoteDatabase) SelectAllQuotes() ([]*QuoteRecord, error) {

	sql := `SELECT id, timestamp, isin, source,
date, price, currency, url, errmsg
FROM quotes
ORDER BY isin, source, date desc
`
	rows, err := qdb.db.Query(sql)
	if err != nil {
		return nil, newError("Select quotes", err)
	}
	defer rows.Close()

	var result []*QuoteRecord
	for rows.Next() {

		r := &QuoteRecord{}
		err = rows.Scan(&r.id, &r.timestamp, &r.isin, &r.source,
			&r.date, &r.price, &r.currency, &r.url, &r.errmsg)
		if err != nil {
			return nil, newError("Select quotes", err)
		}
		result = append(result, r)
	}
	return result, nil
}
*/

/*

// SelectLastQuotes is ...
func (qdb *QuoteDatabase) SelectLastQuotes() ([]*QuoteRecord, error) {

	sqlSelect := `select q.id, q.timestamp, q.isin, q.source,
q.date, q.price, q.currency, q.url, q.errmsg
from quotes q
where q.id in
(
select id
from quotes
where isin = q.isin
and source = q.source
order by timestamp DESC
limit 2
)
order by q.isin, q.timestamp desc, q.source
`
	rows, err := qdb.db.Query(sqlSelect)
	if err != nil {
		return nil, newError("Select last quotes", err)
	}
	defer rows.Close()

	var result []*QuoteRecord
	for rows.Next() {
		var (
			currency, url, errmsg sql.NullString
			price                 sql.NullFloat64
			// date                  sql.NullTime
		)
		r := &QuoteRecord{}
		err = rows.Scan(&r.ID, &r.Timestamp, &r.Isin, &r.Source,
			&r.Date, &price, &currency, &url, &errmsg)
		if err != nil {
			return nil, newError("Select last quotes", err)
		}
		// if date.Valid {
		// 	r.date = date.Time
		// }
		if price.Valid {
			r.Price = float32(price.Float64)
		}
		if currency.Valid {
			r.Currency = currency.String
		}
		if url.Valid {
			r.URL = url.String
		}
		if errmsg.Valid {
			r.ErrMsg = errmsg.String
		}

		result = append(result, r)
	}
	return result, nil
}
*/
