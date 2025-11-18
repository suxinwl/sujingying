package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 打开数据库
	db, err := sql.Open("sqlite3", "./data/sujinying.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 管理员账号信息
	phone := "13900000000"
	password := "admin123"
	role := "super_admin"

	// 检查是否已存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE phone = ?", phone).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	if count > 0 {
		fmt.Printf("⚠️  超级管理员账号已存在: %s\n", phone)
		
		// 显示现有管理员
		fmt.Println("\n当前所有管理员账号:")
		rows, _ := db.Query("SELECT id, phone, role, status FROM users WHERE role IN ('super_admin', 'support')")
		defer rows.Close()
		
		for rows.Next() {
			var id int
			var p, r, s string
			rows.Scan(&id, &p, &r, &s)
			fmt.Printf("  ID: %d, Phone: %s, Role: %s, Status: %s\n", id, p, r, s)
		}
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("密码加密失败:", err)
	}

	// 插入管理员
	result, err := db.Exec(`
		INSERT INTO users (phone, password, role, status, available_deposit, used_deposit, has_pay_password, auto_supplement_enabled, created_at, updated_at)
		VALUES (?, ?, ?, 'active', 0, 0, 0, 0, datetime('now'), datetime('now'))
	`, phone, string(hashedPassword), role)

	if err != nil {
		log.Fatal("创建管理员失败:", err)
	}

	id, _ := result.LastInsertId()
	
	fmt.Println("✅ 超级管理员创建成功！")
	fmt.Println("\n======================")
	fmt.Printf("ID: %d\n", id)
	fmt.Printf("手机号: %s\n", phone)
	fmt.Printf("密码: %s\n", password)
	fmt.Printf("角色: %s\n", role)
	fmt.Println("======================")
	fmt.Println("\n登录地址: http://localhost:5175")
	fmt.Println("\n提示: 使用手机号作为用户名登录")
}
