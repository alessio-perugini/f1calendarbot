package storage

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/alessio-perugini/f1calendarbot/pkg/subscription"
)

type FileStorage struct {
	filePath            string
	subscriptionService subscription.Service
}

func NewFileStorage(
	filePath string,
	subscriptionService subscription.Service,
) *FileStorage {
	return &FileStorage{
		filePath:            filePath,
		subscriptionService: subscriptionService,
	}
}

func (s *FileStorage) DumpSubscribedChats() error {
	var dataToWrite []byte

	for _, v := range s.subscriptionService.GetAllSubscribedChats() {
		dataToWrite = append(dataToWrite, []byte(fmt.Sprintf("%v\n", v))...)
	}

	return os.WriteFile(s.filePath, dataToWrite, os.ModePerm)
}

func (s *FileStorage) LoadSubscribedChats() error {
	buf, err := os.OpenFile(s.filePath, os.O_CREATE|os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}

	defer buf.Close()

	snl := bufio.NewScanner(buf)
	for snl.Scan() {
		chaID, _ := strconv.ParseInt(snl.Text(), 10, 64)
		s.subscriptionService.Subscribe(chaID)
	}

	return snl.Err()
}
