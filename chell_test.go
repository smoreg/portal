package portal

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ifaceless/portal/examples/todo/model"
	"github.com/ifaceless/portal/examples/todo/schema"

	"github.com/stretchr/testify/assert"
)

type NotificationModel struct {
	ID      int
	Title   string
	Content string
}

type UserModel struct {
	ID int
}

func (u *UserModel) Fullname() string {
	return fmt.Sprintf("user:%d", u.ID)
}

func (u *UserModel) Notifications() (result []*NotificationModel) {
	for i := 0; i < 1; i++ {
		result = append(result, &NotificationModel{
			ID:      i,
			Title:   fmt.Sprintf("title_%d", i),
			Content: fmt.Sprintf("content_%d", i),
		})
	}
	return
}

type TaskModel struct {
	ID     int
	UserID int
	Title  string
}

func (t *TaskModel) User() *UserModel {
	return &UserModel{t.UserID}
}

type NotiSchema struct {
	ID      string `json:"id,omitempty"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

type UserSchema struct {
	ID                   string        `json:"id,omitempty"`
	Name                 string        `json:"name,omitempty" portal:"attr:Fullname"`
	Notifications        []*NotiSchema `json:"notifications,omitempty" portal:"nested"`
	AnotherNotifications []*NotiSchema `json:"another_notifications,omitempty" portal:"nested;attr:Notifications"`
}

type TaskSchema struct {
	ID          string      `json:"id,omitempty"`
	Title       string      `json:"title,omitempty"`
	Description string      `json:"description,omitempty" portal:"meth:GetDescription"`
	User        *UserSchema `json:"user,omitempty" portal:"nested;async"`
	SimpleUser  *UserSchema `json:"simple_user,omitempty" portal:"async;nested;only:Name;attr:User"`
}

func (ts *TaskSchema) GetDescription(model *TaskModel) string {
	return "Custom description"
}

func TestDumpOneWithAllFields(t *testing.T) {
	task := TaskModel{
		ID:     1,
		UserID: 1,
		Title:  "Finish your jobs.",
	}

	var taskSchema TaskSchema
	err := Dump(&taskSchema, &task)
	assert.Nil(t, err)

	data, _ := json.Marshal(taskSchema)
	assert.Equal(t, `{"id":"1","title":"Finish your jobs.","description":"Custom description","user":{"id":"1","name":"user:1","notifications":[{"id":"0","title":"title_0","content":"content_0"}],"another_notifications":[{"id":"0","title":"title_0","content":"content_0"}]},"simple_user":{"name":"user:1"}}`, string(data))
}

func TestDumpOneFilterOnlyFields(t *testing.T) {
	task := TaskModel{
		ID:     1,
		UserID: 1,
		Title:  "Finish your jobs.",
	}

	var taskSchema TaskSchema
	err := Dump(&taskSchema, &task, Only("Title", "SimpleUser"))
	assert.Nil(t, err)

	data, _ := json.Marshal(taskSchema)
	assert.Equal(t, `{"title":"Finish your jobs.","simple_user":{"name":"user:1"}}`, string(data))

	var taskSchema2 TaskSchema
	err = Dump(&taskSchema2, &task, Only("ID", "User[ID,Notifications[ID],AnotherNotifications[Title]]", "SimpleUser"))
	assert.Nil(t, err)

	data, _ = json.Marshal(taskSchema2)
	assert.Equal(t, `{"id":"1","user":{"id":"1","notifications":[{"id":"0"}],"another_notifications":[{"title":"title_0"}]},"simple_user":{"name":"user:1"}}`, string(data))
}

func TestDumpOneExcludeFields(t *testing.T) {
	task := TaskModel{
		ID:     1,
		UserID: 1,
		Title:  "Finish your jobs.",
	}

	var taskSchema TaskSchema
	err := Dump(&taskSchema, &task, Exclude("Description", "ID", "User[Name,Notifications[ID,Content],AnotherNotifications], SimpleUser"))
	assert.Nil(t, err)

	data, _ := json.Marshal(taskSchema)
	assert.Equal(t, `{"title":"Finish your jobs.","user":{"id":"1","notifications":[{"title":"title_0"}]}}`, string(data))
}

func TestDumpMany(t *testing.T) {
	var taskSchemas []schema.TaskSchema

	tasks := make([]*model.TaskModel, 0)
	for i := 0; i < 2; i++ {
		tasks = append(tasks, &model.TaskModel{
			ID:     i,
			UserID: i + 100,
			Title:  fmt.Sprintf("Task #%d", i+1),
		})
	}

	err := Dump(&taskSchemas, &tasks, Only("ID", "Title", "User[Name]"))
	assert.Nil(t, err)

	data, _ := json.Marshal(taskSchemas)
	assert.Equal(t, `[{"id":"0","title":"Task #1","user":{"name":"user:100"}},{"id":"1","title":"Task #2","user":{"name":"user:101"}}]`, string(data))
}

func TestDumpError(t *testing.T) {
	task := TaskModel{
		ID:     1,
		UserID: 1,
		Title:  "Finish your jobs.",
	}

	var taskSchema TaskSchema
	err := Dump(&taskSchema, &task, Only("Title", "SimpleUser["))
	assert.NotNil(t, err)
	assert.Equal(t, ErrUnmatchedBrackets.Error(), err.Error())
}

func TestChellDumpOk(t *testing.T) {
	task := TaskModel{
		ID:     1,
		UserID: 1,
		Title:  "Finish your jobs.",
	}

	var taskSchema TaskSchema
	chell := New().Only("Title", "SimpleUser")
	err := chell.Dump(&taskSchema, &task)
	assert.Nil(t, err)

	data, _ := json.Marshal(taskSchema)
	assert.Equal(t, `{"title":"Finish your jobs.","simple_user":{"name":"user:1"}}`, string(data))

	chell = New().Exclude("Description", "ID", "User[Name,Notifications[ID,Content],AnotherNotifications], SimpleUser")
	var taskSchema2 TaskSchema
	err = chell.Dump(&taskSchema2, &task)
	assert.Nil(t, err)

	data, _ = json.Marshal(taskSchema2)
	assert.Equal(t, `{"title":"Finish your jobs.","user":{"id":"1","notifications":[{"title":"title_0"}]}}`, string(data))
}

func TestChellBoundaryConditions(t *testing.T) {
	task := TaskModel{
		ID:     1,
		UserID: 1,
		Title:  "Finish your jobs.",
	}

	var taskSchema TaskSchema
	err := Dump(taskSchema, task)
	assert.NotNil(t, err)
	assert.Equal(t, "dst must be a pointer", err.Error())
}
