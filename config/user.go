package config

import (
	"dxkite.cn/log"
	"encoding/json"
	"os"
	"strconv"
	"time"
)

type UserConfig struct {
	Enable   bool       `yaml:"enable"`
	UserList []UserItem `yaml:"user_list"`
}

type UserItem struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Address  string `yaml:"address"`
	Freeze   bool   `yaml:"freeze"`
}

type UserFlowJSON struct {
	UserId string `json:"user_id"`
	Up     string `json:"up"`
	Down   string `json:"down"`
}

type UserFlow struct {
	UserId string
	Up     int64
	Down   int64
}

type UserFlowMap map[string]*UserFlow

func (uf *UserFlowMap) Save(p string) error {
	ufj := make([]UserFlowJSON, len(*uf))
	i := 0
	for _, v := range *uf {
		ufj[i] = UserFlowJSON{
			UserId: v.UserId,
			Up:     strconv.Itoa(int(v.Up)),
			Down:   strconv.Itoa(int(v.Down)),
		}
	}
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	e := json.NewEncoder(f)
	if err := e.Encode(ufj); err != nil {
		return err
	}
	log.Info("user flow write success", p)
	return nil
}

func (uf *UserFlowMap) Update(user string, up, down int64) {
	if _, ok := (*uf)[user]; ok {
		(*uf)[user].Up += up
		(*uf)[user].Down += down
	} else {
		(*uf)[user] = &UserFlow{
			UserId: user,
			Up:     up,
			Down:   down,
		}
	}
}

func (uf *UserFlowMap) Write(p string, duration int) {
	log.Info("async write", p)
	ticker := time.NewTicker(time.Duration(duration) * time.Second)
	for range ticker.C {
		if err := uf.Save(p); err != nil {
			log.Error("write file error", err)
		}
	}
}

func (uf *UserFlowMap) Load(p string) error {
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	d := json.NewDecoder(f)
	ufj := []UserFlowJSON{}
	if err := d.Decode(&ufj); err != nil {
		return err
	}
	for _, v := range ufj {
		down, _ := strconv.Atoi(v.Down)
		up, _ := strconv.Atoi(v.Up)
		(*uf)[v.UserId] = &UserFlow{
			UserId: v.UserId,
			Up:     int64(up),
			Down:   int64(down),
		}
	}
	return nil
}
