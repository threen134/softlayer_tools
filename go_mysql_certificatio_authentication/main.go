package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// path to cert-files hard coded
// Most of this is copy pasted from the internet
// and used without much reflection
func createTLSConf() tls.Config {

	rootCertPool := x509.NewCertPool()
	pem, err := ioutil.ReadFile("ca-cert.pem")
	if err != nil {
		log.Fatal(err)
	}
	if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
		log.Fatal("Failed to append PEM.")
	}
	// clientCert := make([]tls.Certificate, 0, 1)

	// certs, err := tls.LoadX509KeyPair("cert/client-cert.pem", "cert/client-key.pem")
	// if err != nil {
	//     log.Fatal(err)
	// }

	// clientCert = append(clientCert, certs)

	return tls.Config{
		RootCAs: rootCertPool,
		//    Certificates:       clientCert,
		InsecureSkipVerify: true, // needed for self signed certs
	}
}

// Test that db is usable
// prints version to stdout
func queryDB(db *sql.DB) {
	// Query the database
	var result string
	err := db.QueryRow("SELECT NOW()").Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}

func createProductTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS product(product_id int primary key auto_increment, product_name text, 
        product_price int, created_at datetime default CURRENT_TIMESTAMP, updated_at datetime default CURRENT_TIMESTAMP)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when creating product table", err)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when getting rows affected", err)
		return err
	}
	log.Printf("Rows affected when creating table: %d", rows)
	return nil
}

func main() {

	dbHost := "2eac3d20-c275-46fc-a24b-4a993fc0dcdb.c5kn1n9d0g7polghe820.databases.appdomain.cloud"
	dbPort := "30931"
	username := "916f6958567147fca6d0fe4fdbdbccdb"
	password := "···"

	// When I realized that the tls/ssl/cert thing was handled separately
	// it became easier, the following two lines are the important bit
	tlsConf := createTLSConf()
	err := mysql.RegisterTLSConfig("custom", &tlsConf)
	
	if err != nil {
		log.Printf("Error %s when RegisterTLSConfig\n", err)
		return
	}

	// connection string (dataSourceName) is slightly different
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=custom", username, password, dbHost, dbPort, "ibmclouddb")
	db1, err := sql.Open("mysql", dsn)

	if err != nil {
		log.Printf("Error %s when opening DB\n", err)
		log.Printf("%s", dsn)
		return
	}

	err = createProductTable(db1)
	if err != nil {
		log.Printf("Create product table failed with error %s", err)
		return
	}

	defer db1.Close()
	e := db1.Ping()
	fmt.Println(dsn, e)
	queryDB(db1)
}
