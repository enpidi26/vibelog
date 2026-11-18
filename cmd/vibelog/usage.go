package main

import "fmt"

func usage() {
	fmt.Print(`vibelog — prompt journal

Usage:
  vibelog              start the server (env-configured)
  vibelog backup [dir] write a consistent SQLite backup into dir
  vibelog version      print the version
  vibelog help         show this message

Environment:
  BASE_URL                 public URL, e.g. https://vibelog.example.com
  DATABASE_PATH            SQLite file (default data/vibelog.db)
  JWT_SECRET               optional; otherwise generated next to the database
  ALLOW_REGISTRATION       true|false (default true)
  EMAIL_DRIVER             console|smtp|resend|off (default console)
  SMTP_HOST/SMTP_PORT/...  SMTP settings when EMAIL_DRIVER=smtp
  RESEND_API_KEY           when EMAIL_DRIVER=resend
`)
}
