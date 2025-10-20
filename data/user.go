package data

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Note: User struct is now defined in models.go

// UserRepository implements UserInterface using GORM.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository.
func NewUserRepository(db *gorm.DB) UserInterface {
	return &UserRepository{db: db}
}

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// GetAll retrieves all users
func (u *UserRepository) GetAll() ([]*User, error) {
	var users []*User
	result := u.db.Find(&users)
	return users, result.Error
}

// GetByEmail retrieves a user by email
func (u *UserRepository) GetByEmail(email string) (*User, error) {
	var user User
	result := u.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetOne retrieves a user by ID
func (u *UserRepository) GetOne(id uint) (*User, error) {
	var user User
	result := u.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// Insert creates a new user
func (u *UserRepository) Insert(user *User) (uint, error) {
	// Hash the password before saving
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return 0, err
	}
	user.Password = hashedPassword

	result := u.db.Create(user)
	return user.ID, result.Error
}

// Update updates an existing user
func (u *UserRepository) Update(user *User) error {
	// If password is being updated, hash it
	if user.Password != "" {
		hashedPassword, err := HashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	}

	result := u.db.Save(user)
	return result.Error
}

// Delete soft deletes a user
func (u *UserRepository) Delete(user *User) error {
	result := u.db.Delete(user)
	return result.Error
}

// DeleteByID soft deletes a user by ID
func (u *UserRepository) DeleteByID(id uint) error {
	result := u.db.Delete(&User{}, id)
	return result.Error
}

// ResetPassword resets a user's password
func (u *UserRepository) ResetPassword(userID uint, newPassword string) error {
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	result := u.db.Model(&User{}).Where("id = ?", userID).Update("password", hashedPassword)
	return result.Error
}

// PasswordMatches checks if the provided password matches the user's password
func (u *UserRepository) PasswordMatches(user *User, plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(plainText))
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GenerateAndSaveOTP generates a 6-digit OTP and saves it to the user
func (u *UserRepository) GenerateAndSaveOTP(email string) (string, error) {
	// Generate 6-digit OTP
	otp, err := generateOTP()
	if err != nil {
		return "", err
	}

	// Set expiration time (10 minutes from now)
	expiresAt := time.Now().Add(10 * time.Minute)

	// Update user with OTP and expiration
	result := u.db.Model(&User{}).Where("email = ?", email).Updates(map[string]interface{}{
		"otp_code":       otp,
		"otp_expires_at": expiresAt,
	})

	if result.Error != nil {
		return "", result.Error
	}

	return otp, nil
}

// VerifyOTP verifies if the provided OTP is valid for the email
func (u *UserRepository) VerifyOTP(email, otp string) (bool, error) {
	var user User
	result := u.db.Where("email = ? AND otp_code = ? AND otp_expires_at > ?",
		email, otp, time.Now()).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, result.Error
	}

	return true, nil
}

// ResetPasswordWithOTP resets password using OTP verification
func (u *UserRepository) ResetPasswordWithOTP(email, otp, newPassword string) error {
	// First verify the OTP
	valid, err := u.VerifyOTP(email, otp)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("invalid or expired OTP")
	}

	// Hash the new password
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password and clear OTP
	result := u.db.Model(&User{}).Where("email = ?", email).Updates(map[string]interface{}{
		"password":       hashedPassword,
		"otp_code":       "",
		"otp_expires_at": nil,
	})

	return result.Error
}

// generateOTP generates a random 6-digit OTP
func generateOTP() (string, error) {
	// Generate a random number between 100000 and 999999
	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}
