package globals

import (
	"context"
)

var (
	// tokenSigningAlgo = jwt.SigningMethodHS256
	JwtSecret = []byte("your_secret_key") // Replace with a secure secret key
)

// Context keys
type ContextKey string

const RoleKey ContextKey = "role"
const UserIDKey ContextKey = "userId"

var Ctx = context.Background()
