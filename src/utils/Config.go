package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	dbClient "localapps-server/db/client"

	db "localapps-server/db/generated"
	"localapps-server/types"
	"reflect"
)

var ServerConfig types.ServerConfig

func UpdateServerConfigCache() error {
	err := validateServerConfig()
	if err != nil {
		return err
	}

	client, _ := dbClient.GetClient()
	config, err := client.GetConfig(context.Background())
	if err != nil {
		return err
	}

	configMap := make(map[string]string)
	for _, c := range config {
		configMap[c.Key] = c.Value.String
	}

	configType := reflect.TypeOf(ServerConfig)
	for i := range configType.NumField() {
		field := configType.Field(i)

		if _, ok := configMap[field.Name]; ok {
			fieldValue := reflect.ValueOf(&ServerConfig).Elem().FieldByName(field.Name)
			json.Unmarshal([]byte(configMap[field.Name]), fieldValue.Addr().Interface())
		}
	}
	return nil
}

func validateServerConfig() error {
	client, _ := dbClient.GetClient()
	config, err := client.GetConfig(context.Background())
	if err != nil {
		return err
	}

	configStruct := reflect.TypeOf(types.ServerConfig{})
	var missingKeys []string

	configMap := make(map[string]string)
	for _, c := range config {
		configMap[c.Key] = c.Value.String
	}

	for i := range configStruct.NumField() {
		field := configStruct.Field(i)

		if _, ok := configMap[field.Name]; !ok {
			missingKeys = append(missingKeys, field.Name)
		}
	}

	for _, k := range missingKeys {
		field, _ := reflect.TypeOf(types.ServerConfig{}).FieldByName(k)
		defaultValue := field.Tag.Get("default")

		client.SetConfigKey(context.Background(), db.SetConfigKeyParams{Key: k, Value: sql.NullString{String: defaultValue, Valid: true}})
	}
	return nil
}
