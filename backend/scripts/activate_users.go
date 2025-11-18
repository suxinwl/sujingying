package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// 打开数据库
	db, err := sql.Open("sqlite3", "./data/sujinying.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 激活所有待审核用户
	result, err := db.Exec("UPDATE users SET status = 'active' WHERE status = 'pending'")
	if err != nil {
		log.Fatal(err)
	}
	
	rows, _ := result.RowsAffected()
	fmt.Printf("✅ Activated %d users\n", rows)
	
	// 显示所有用户
	fmt.Println("\nUsers in database:")
	stmt, err := db.Query("SELECT id, phone, role, status FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	
	for stmt.Next() {
		var id int
		var phone, role, status string
		stmt.Scan(&id, &phone, &role, &status)
		fmt.Printf("  ID: %d, Phone: %s, Role: %s, Status: %s\n", id, phone, role, status)
	}
	
	fmt.Println("\n✅ All users activated successfully!")
	fmt.Println("\nYou can now login with:")
	fmt.Println("  - admin/123456")
	fmt.Println("  - customer/123456")
}
