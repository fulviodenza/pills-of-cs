// Code generated by ent, DO NOT EDIT.

package user

const (
	// Label holds the string label denoting the user type in the database.
	Label = "user"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "user_id"
	// FieldCategories holds the string denoting the categories field in the database.
	FieldCategories = "categories"
	// FieldPillSchedule holds the string denoting the pill_schedule field in the database.
	FieldPillSchedule = "pill_schedule"
	// FieldNewsSchedule holds the string denoting the news_schedule field in the database.
	FieldNewsSchedule = "news_schedule"
	// Table holds the table name of the user in the database.
	Table = "users"
)

// Columns holds all SQL columns for user fields.
var Columns = []string{
	FieldID,
	FieldCategories,
	FieldPillSchedule,
	FieldNewsSchedule,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}
