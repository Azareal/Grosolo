/* Copyright Azareal 2017 - 2018 */
package main

import (
	"fmt"
	"os"
	"bufio"
	"strconv"
	"database/sql"
	"runtime/debug"
	
	"../query_gen/lib"
)

const saltLength int = 32
var db *sql.DB
var scanner *bufio.Scanner

var db_adapter string = "mysql"
var db_host string
var db_username string
var db_password string
var db_name string
var db_port string
var site_name, site_url, server_port string

var default_adapter string = "mysql"
var default_host string = "localhost"
var default_username string = "root"
var default_dbname string = "gosora"
var default_site_name string = "Site Name"
var default_site_url string = "localhost"
var default_server_port string = "80" // 8080's a good one, if you're testing and don't want it to clash with port 80

var init_database func()error = _init_mysql
var table_defs func()error = _table_defs_mysql
var initial_data func()error = _initial_data_mysql

func main() {
	// Capture panics rather than immediately closing the window on Windows
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println(r)
			debug.PrintStack()
			press_any_key()
			return
		}
	}()
	
	scanner = bufio.NewScanner(os.Stdin)
	fmt.Println("Welcome to Gosora's Installer")
	fmt.Println("We're going to take you through a few steps to help you get started :)")
	if !get_database_details() {
		err := scanner.Err()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Something went wrong!")
		}
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}

	if !get_site_details() {
		err := scanner.Err()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Something went wrong!")
		}
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	err := init_database()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	err = table_defs()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	hashed_password, salt, err := BcryptGeneratePassword("password")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	// Build the admin user query
	admin_user_stmt, err := qgen.Builder.SimpleInsert("users","name, password, salt, email, group, is_super_admin, active, createdAt, lastActiveAt, message, last_ip","'Admin',?,?,'admin@localhost',1,1,1,NOW(),NOW(),'','127.0.0.1'")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	// Run the admin user query
	_, err = admin_user_stmt.Exec(hashed_password,salt)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	err = initial_data()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}
	
	if db_adapter == "mysql" {
		err = _mysql_seed_database()
		if err != nil {
			fmt.Println(err)
			fmt.Println("Aborting installation...")
			press_any_key()
			return
		}
	}
	
	configContents := []byte(`package main

// Site Info
var site_name = "` + site_name + `" // Should be a setting in the database
var site_url = "` + site_url + `"
var server_port = "` + server_port + `"
var enable_ssl = false
var ssl_privkey = ""
var ssl_fullchain = ""

// Database details
var dbhost = "` + db_host + `"
var dbuser = "` + db_username + `"
var dbpassword = "` + db_password + `"
var dbname = "` + db_name + `"
var dbport = "` + db_port + `" // You probably won't need to change this

// Limiters
var max_request_size = 5 * megabyte

// Caching
var cache_topicuser = CACHE_STATIC
var user_cache_capacity = 100 // The max number of users held in memory
var topic_cache_capacity = 100 // The max number of topics held in memory

// Email
var site_email = "" // Should be a setting in the database
var smtp_server = ""
var smtp_username = ""
var smtp_password = ""
var smtp_port = "25"
var enable_emails = false

// Misc
var default_route = route_topics
var default_group = 3 // Should be a setting in the database
var activation_group = 5 // Should be a setting in the database
var staff_css = " background-color: #ffeaff;"
var uncategorised_forum_visible = true
var minify_templates = true
var multi_server = false // Experimental: Enable Cross-Server Synchronisation and several other features

//var noavatar = "https://api.adorable.io/avatars/{width}/{id}@{site_url}.png"
var noavatar = "https://api.adorable.io/avatars/285/{id}@" + site_url + ".png"
var items_per_page = 25

// Developer flag
var debug_mode = true
var super_debug = false
var profiling = false
`)

	fmt.Println("Opening the configuration file")
	configFile, err := os.Create("./config.go")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}

	fmt.Println("Writing to the configuration file...")
	_, err = configFile.Write(configContents)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		press_any_key()
		return
	}

	configFile.Sync()
	configFile.Close()
	fmt.Println("Finished writing to the configuration file")

	fmt.Println("Yay, you have successfully installed Gosora!")
	fmt.Println("Your name is Admin and you can login with the password 'password'. Don't forget to change it! Seriously. It's really insecure.")
	press_any_key()
}

func get_database_details() bool {
	fmt.Println("Which database driver do you wish to use? mysql, mysql, or mysql? Default: mysql")
	if !scanner.Scan() {
		return false
	}
	db_adapter = scanner.Text()
	if db_adapter == "" {
		db_adapter = default_adapter
	}
	db_adapter = set_db_adapter(db_adapter)
	fmt.Println("Set database adapter to " + db_adapter)
	
	fmt.Println("Database Host? Default: " + default_host)
	if !scanner.Scan() {
		return false
	}
	db_host = scanner.Text()
	if db_host == "" {
		db_host = default_host
	}
	fmt.Println("Set database host to " + db_host)

	fmt.Println("Database Username? Default: " + default_username)
	if !scanner.Scan() {
		return false
	}
	db_username = scanner.Text()
	if db_username == "" {
		db_username = default_username
	}
	fmt.Println("Set database username to " + db_username)

	fmt.Println("Database Password? Default: ''")
	if !scanner.Scan() {
		return false
	}
	db_password = scanner.Text()
	if len(db_password) == 0 {
		fmt.Println("You didn't set a password for this user. This won't block the installation process, but it might create security issues in the future.\n")
	} else {
		fmt.Println("Set password to " + obfuscate_password(db_password))
	}

	fmt.Println("Database Name? Pick a name you like or one provided to you. Default: " + default_dbname)
	if !scanner.Scan() {
		return false
	}
	db_name = scanner.Text()
	if db_name == "" {
		db_name = default_dbname
	}
	fmt.Println("Set database name to " + db_name)
	return true
}

func get_site_details() bool {
	fmt.Println("Okay. We also need to know some actual information about your site!")
	fmt.Println("What's your site's name? Default: " + default_site_name)
	if !scanner.Scan() {
		return false
	}
	site_name = scanner.Text()
	if site_name == "" {
		site_name = default_site_name
	}
	fmt.Println("Set the site name to " + site_name)

	fmt.Println("What's your site's url? Default: " + default_site_url)
	if !scanner.Scan() {
		return false
	}
	site_url = scanner.Text()
	if site_url == "" {
		site_url = default_site_url
	}
	fmt.Println("Set the site url to " + site_url)

	fmt.Println("What port do you want the server to listen on? If you don't know what this means, you should probably leave it on the default. Default: " + default_server_port)
	if !scanner.Scan() {
		return false
	}
	server_port = scanner.Text()
	if server_port == "" {
		server_port = default_server_port
	}
	_, err := strconv.Atoi(server_port)
	if err != nil {
		fmt.Println("That's not a valid number!")
		return false
	}
	fmt.Println("Set the server port to " + server_port)
	return true
}

func set_db_adapter(name string) string {
	switch(name) {
		//case "wip-pgsql":
		//	set_pgsql_adapter()
		//	return "wip-pgsql"	
	}
	_set_mysql_adapter()
	return "mysql"
}

func obfuscate_password(password string) (out string) {
	for i := 0; i < len(password); i++ {
		out += "*"
	}
	return out
}

func press_any_key() {
	//fmt.Println("Press any key to exit...")
	fmt.Println("Please press enter to exit...")
	for scanner.Scan() {
		_ = scanner.Text()
		return
	}
}
