package models

type User struct {
    ID       int64  `json:"id"`
    Username string `json:"username"`
    Password string `json:"password"`
    Email    string `json:"email"`
}

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}
