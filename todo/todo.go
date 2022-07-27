package todo

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/alexeyco/simpletable"
)

type Todo struct {
	Task 			string 		`json:"task"`
	Done 			bool 		`json:"done"`
	Created_at 		time.Time 	`json:"created_at"`
	Completed_at 	time.Time 	`json:"completed_at"`
	Updated_at 		time.Time 	`json:"updated_at"`
}

type Todos []Todo

const (
	todoFile = "todos.json"
)

func TodoMain() {
	add := flag.Bool("add", false, "add a new todo")
	complete := flag.Int("complete", 0, "mark a todo as completed")
	del := flag.Int("del", 0, "delete a todo")
	list := flag.Bool("list", false, "all list")
	update := flag.Int("update", 0, "update a todo")

	flag.Parse()

	todos := &Todos{}
	if err := todos.load(todoFile); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	switch {
	case *add:
		task, err := todos.getInput(os.Stdin, flag.Args()...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		todos.add(task)
		err = todos.store(todoFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case *complete > 0:
		err := todos.complete(*complete)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		err = todos.store(todoFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case *del > 0:
		err := todos.delete(*del)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		err = todos.store(todoFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case *list:
		todos.print(todoFile)
	case *update > 0:
		_, err := todos.update(*update, os.Stdin, flag.Args()...)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		err = todos.store(todoFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stdout, "invalid command")
		os.Exit(0)
	}
}

func (t *Todos) add(task string) {
	todo := Todo{
		Task: task,
		Done: false,
		Created_at: time.Now(),
		Completed_at: time.Time{},
		Updated_at: time.Time{},
	}
	*t = append(*t, todo)
}

func (t *Todos) complete(i int) error {
	todo := *t
	if i <= 0 || i > len(todo) {
		return errors.New("invalid index")
	}
	todo[i-1].Done = true
	todo[i-1].Completed_at = time.Now()
	return nil
}

func (t *Todos) delete(i int) error {
	todo := *t
	if i <= 0 || i >= len(todo) {
		return errors.New("invalid index")
	}
	*t = append(todo[:i-1], todo[i:]...)
	return nil
}

func (t *Todos) update(i int, r io.Reader, args ...string) (string, error) {
	todo := *t
	if i <= 0 || i > len(todo) {
		return "", errors.New("invalid index")
	}
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	fmt.Print("Enter task: ")
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", err
	}
	text := scanner.Text()
	if len(text) == 0 {
		return "", errors.New("empty input is not allowed")
	}
	todo[i-1].Task = text
	todo[i-1].Done = true
	todo[i-1].Completed_at = time.Time{}
	todo[i-1].Updated_at = time.Now()

	return text, nil
}

func (t *Todos) print(filename string) {
	table := simpletable.New()
	table.Header = &simpletable.Header{
		Cells:  []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: Gray("#")},
			{Align: simpletable.AlignCenter, Text: Red("Task")},
			{Align: simpletable.AlignCenter, Text: Blue("Done")},
			{Align: simpletable.AlignCenter, Text: Green("Created_at")},
			{Align: simpletable.AlignCenter, Text: Green("Completed_at")},
			{Align: simpletable.AlignCenter, Text: Green("Updated_at")},
		},
	}

	var pending int
	for i, v := range *t {
		done := fmt.Sprintf("%t", v.Done)
		if done == "true" {
			done = Green("Yes")
		} else {
			done = Blue("No")
			pending++
		}
		r := []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: fmt.Sprintf("%d", i)},
			{Text: Red(v.Task)},
			{Align: simpletable.AlignCenter, Text: done},
			{Align: simpletable.AlignCenter, Text: Green(fmt.Sprint(v.Created_at.Format("01-02-2006 15:04:05 Mon")))},
			{Align: simpletable.AlignCenter, Text: Green(fmt.Sprint(v.Completed_at.Format("01-02-2006 15:04:05 Mon")))},
			{Align: simpletable.AlignCenter, Text: Green(fmt.Sprint(v.Updated_at.Format("01-02-2006 15:04:05 Mon")))},
		}
		table.Body.Cells = append(table.Body.Cells, r)
	}
	table.Footer = &simpletable.Footer{
		Cells: []*simpletable.Cell{
			{},
			{},
			{Align: simpletable.AlignCenter, Text: fmt.Sprintf("pending %d tasks", pending)},
			{},
			{},
			{},
		},
	}

	table.SetStyle(simpletable.StyleCompactLite)
	fmt.Println(table.String())
}

func (t *Todos) store(filename string) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func (t *Todos) load(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, t)
	if err != nil {
		return err
	}
	return nil
}

// command
// go run main.go -add task
// echo "task" | ./main -add
func (t *Todos) getInput(r io.Reader, args ...string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	fmt.Print("Enter task: ")
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", err
	}
	text := scanner.Text()
	if len(text) == 0 {
		return "", errors.New("empty input is not allowed")
	}
	return text, nil
}

