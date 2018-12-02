package tikv

import (
	"github.com/pingcap/tidb/config"
	"github.com/pingcap/tidb/store/tikv"
	"log"
)

type kv struct {
	Key   string
	Value string
}

var Client *tikv.RawKVClient

// InitStore
func InitStore() {
	var err error
	Client, err = tikv.NewRawKVClient([]string{"117.50.61.185:2379"}, config.Security{})
	if err != nil {
		log.Println(err)
	}
}

// Puts
func Puts(args ...[]byte) error {
	//log.Printf("%v",args)
	for i := 0; i < len(args); i += 2 {
		key, val := args[i], args[i+1]
		err := Client.Put(key, val)
		if err != nil {
			return err
		}
	}
	return nil
}

// Dels
func Dels(keys ...[]byte) error {
	for i := 0; i < len(keys); i += 1 {
		err := Client.Delete(keys[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// Delall
func Delall(startKey []byte, limit int) error {
	keys, _, err := Client.Scan(startKey, limit)
	if err != nil {
		return err
	}
	for i := 0; i < len(keys); i += 1 {
		Dels(keys[i])
	}
	return nil
}

// Get
func Get(k []byte) (kv, error) {
	v, err := Client.Get(k)
	if err != nil {
		return kv{}, err
	}
	return kv{Key: string(k), Value: string(v)}, nil
}

// Scan
func Scan(startKey []byte, limit int) ([]kv, error) {
	var kvs []kv
	keys, values, err := Client.Scan(startKey, limit)
	if err != nil {
		return kvs, err
	}
	for i := 0; i < len(keys); i += 1 {
		kvs = append(kvs, kv{Key: string(keys[i]), Value: string(values[i])})
	}
	return kvs, nil
}
