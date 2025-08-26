package middleware

import (
    "errors"
    "fmt"
    "net/http"
    "strings"

    "github.com/golang-jwt/jwt/v5"
    "github.com/labstack/echo/v4"
    "gorm.io/gorm"
)

type SupabaseClaims struct {
    Role  string `json:"role"`
    Email string `json:"email"`
    jwt.RegisteredClaims
}

// NewAuthMiddleware verifies Supabase JWT (HS256) from Authorization: Bearer <token>,
// ensures a local User exists (auto-provision if missing), and injects user_id/user_role.
func NewAuthMiddleware(db *gorm.DB, jwtSecret string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            authz := c.Request().Header.Get("Authorization")
            if authz == "" || !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
                return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid Authorization header")
            }

            tokenStr := strings.TrimSpace(authz[len("Bearer "):])
            claims := &SupabaseClaims{}

            token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
                if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"]) 
                }
                return []byte(jwtSecret), nil
            })
            if err != nil || !token.Valid {
                return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
            }

            // sub is the Supabase Auth user UUID
            supabaseUID := claims.Subject
            if supabaseUID == "" {
                return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token: missing subject")
            }

            // Ensure a local user exists and retrieve role
            var result struct {
                ID   uint
                Role string
            }

            err = db.Raw("SELECT id, role FROM users WHERE supabase_uid = ? LIMIT 1", supabaseUID).Scan(&result).Error
            if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && result.ID == 0) {
                email := claims.Email
                if email == "" {
                    email = fmt.Sprintf("user_%s@local", supabaseUID)
                }
                if err := db.Exec("INSERT INTO users (email, name, role, supabase_uid, created_at, updated_at) VALUES (?, ?, 'employee', ?, NOW(), NOW())",
                    email, email, supabaseUID).Error; err != nil {
                    return echo.NewHTTPError(http.StatusUnauthorized, "Failed to provision user")
                }
                if err := db.Raw("SELECT id, role FROM users WHERE supabase_uid = ? LIMIT 1", supabaseUID).Scan(&result).Error; err != nil || result.ID == 0 {
                    return echo.NewHTTPError(http.StatusUnauthorized, "Failed to load user")
                }
            } else if err != nil {
                return echo.NewHTTPError(http.StatusUnauthorized, "Auth error")
            }

            c.Set("user_id", result.ID)
            c.Set("user_role", result.Role)

            return next(c)
        }
    }
}

func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        roleVal := c.Get("user_role")
        role, _ := roleVal.(string)
        if role != "admin" && role != "manager" {
            return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
        }
        return next(c)
    }
}
