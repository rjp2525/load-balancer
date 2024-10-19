package balancer

import (
    "encoding/json"
    "io/ioutil"
)

type Config struct {
    Port    string   `json:"port"`
    Servers []string `json:"servers"`
}

// Load config from a JSON file
func LoadConfig(filename string) (*Config, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var config Config
    err = json.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }

    return &config, nil
}
