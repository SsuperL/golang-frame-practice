package session

import (
	"fmt"
	"testing"

	"github.com/go-playground/assert"
)

func initTest(t *testing.T) *Session {
	t.Helper()
	session := NewSession()
	session = session.Model(&User{})
	if session.HasTable("User") {
		session.DropTable()
	}
	err2 := session.CreateTable()
	if err2 != nil {
		err := fmt.Errorf("err2:%v", err2)
		t.Fatalf("failed to init : %v", err)
	}
	session.Raw("INSERT INTO User VALUES (?,?),(?,?)", "Tom", 12, "John", 14).Exec()
	return session
}
func TestSession_Insert(t *testing.T) {
	session := initTest(t)
	rowsAffected, err := session.Insert(&User{"April", 15})
	if err != nil {
		t.Fatal("Insert failed")
	}
	assert.Equal(t, 1, int(rowsAffected))
}

func TestSession_Find(t *testing.T) {
	session := initTest(t)
	var users []User
	err := session.Find(&users)
	if err != nil {
		t.Fatal("Find failed")
	}
	assert.Equal(t, 2, len(users))
}

func TestSession_Update(t *testing.T) {
	session := initTest(t)
	newAge := 22
	affectedRow, err := session.Where("Name = ?", "Tom").Update(map[string]interface{}{"Age": newAge})
	if err != nil {
		t.Fatal("Update failed")
	}
	assert.Equal(t, 1, int(affectedRow))

	row := session.Raw("SELECT  Age FROM User").Where("Name", "Tom").QueryRow()
	var age int
	if err := row.Scan(&age); err != nil {
		t.Fatalf(fmt.Sprintf("Scan failed: %v", err))
	}
	assert.Equal(t, newAge, int(age))

	ageOfJohn := 16
	affectedRow, err = session.Where("Name = ?", "John").Update("Age", ageOfJohn)
	if err != nil {
		t.Fatalf(fmt.Sprintf("Update failed: %v", err))
	}
	assert.Equal(t, 1, int(affectedRow))
}

func TestSession_Count(t *testing.T) {
	session := initTest(t)
	count, err := session.Where("Name = ?", "Tom").Count()
	if err != nil {
		t.Fatalf(fmt.Sprintf("Count failed: %v", err))
	}
	assert.Equal(t, 1, int(count))
}

func TestSession_Limit(t *testing.T) {
	session := initTest(t)
	var users []User
	err := session.Limit(1).Find(&users)
	if err != nil {
		t.Fatalf("Limit failed: %v", err)
	}
	assert.Equal(t, 1, len(users))

}
func TestSession_First(t *testing.T) {
	session := initTest(t)
	var user User
	err := session.OrderBy("Name").First(&user)
	if err != nil {
		t.Fatalf("Get first record failed: %v", err)
	}
	assert.Equal(t, "John", user.Name)
}

func TestSession_Delete(t *testing.T) {
	session := initTest(t)
	rowAffected, err := session.Where("Name = ?", "Tom").Delete("User")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	assert.Equal(t, 1, int(rowAffected))
	var users []User
	_ = session.Find(&users)
	assert.Equal(t, 1, len(users))
}
