package models

type User struct {
    ID           int    `db:"id"`
    FirstName    string `db:"first_name"`
    LastName     string `db:"last_name"`
    Email        string `db:"email"`
    Username     string `db:"username"`
    PasswordHash string `db:"password_hash"`
    TwoFAEnabled bool   `db:"two_fa_enabled"`
}