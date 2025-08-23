package greetings

import "errors"

// Hello returns a greeting for the named person.
func Hello(name string) (string, error) {
    if name == "" {
        return "", errors.New("empty name")
    }
    message := "Hi, " + name + ". Welcome!"
    return message, nil
}
