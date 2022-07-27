package todo

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type Todo struct {
	Task string
	Done bool
	Created_at time.Time
	Completed_at time.Time
}

type Todos []Todo

const (
	todoFile = "todos.json"
)

func TodoMain() {
	add := flag.Bool("add", false, "add a new todo")
	flag.Parse()

	todos := &Todos{}
	if err := todos.load(todoFile); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	switch {
	case *add:
		todos.add("sample todo")
		err := todos.store(todoFile)
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
	}

	*t = append(*t, todo)
}

func (t *Todos) complete(i int) error {
	todo := *t
	if i <= 0 || i >= len(todo) {
		return errors.New("invalid index")
	}

	todo[i].Done = true
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

func (t *Todos) update(v string, i int) error {
	todo := *t
	if i <= 0 || i >= len(todo) {
		return errors.New("invalid index")
	}
	todo[i-1] = Todo{
		Task: v,
		Done: false,
		Created_at: time.Now(),
		Completed_at: time.Time{},
	}

	return nil
}

func (t *Todos) load(filename string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	if len(file) == 0 {
		return err
	}
	err = json.Unmarshal(file, t)
	if err != nil {
		return err
	}

	return nil
}

func (t *Todos) store(filename string) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, 0644)
}