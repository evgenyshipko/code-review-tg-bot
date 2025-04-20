package stories

import "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type ExecutionStatus string

type StageCallback func(update tgbotapi.Update, s *StoryService) ExecutionStatus

const (
	Executed    ExecutionStatus = "executed"
	Dropped     ExecutionStatus = "dropped"
	NotExecuted ExecutionStatus = "not_executed"
	Finished    ExecutionStatus = "finished"
)

type Stage struct {
	cb   StageCallback
	name string
}

type Story struct {
	name   string
	stages []Stage
}

func (s *Story) getStage(name string) *Stage {
	for _, stage := range s.stages {
		if stage.name == name {
			return &stage
		}
	}
	return nil
}

func (s *Story) getNextStage(prevName string) *Stage {
	nextIndex := 0
	for index, stage := range s.stages {
		if stage.name == prevName {
			nextIndex = index + 1
			break
		}
	}
	if nextIndex >= len(s.stages) {
		return nil
	}
	return &s.stages[nextIndex]
}

func NewStage(name string, cb StageCallback) *Stage {
	return &Stage{name: name, cb: cb}
}

func NewStory(name string, stages []Stage) *Story {
	return &Story{
		name:   name,
		stages: stages,
	}
}
