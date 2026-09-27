package main
import (
    "fmt"
    "reflect"
    "github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
)
func main() {
    u := models.User{}
    t := reflect.TypeOf(u)
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        fmt.Println(field.Name, field.Tag.Get("gorm"))
    }
}
