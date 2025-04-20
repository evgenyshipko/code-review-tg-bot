package stories

import (
	"code-review-tg-bot/internal/access"
	"code-review-tg-bot/internal/logger"
	"code-review-tg-bot/internal/storage"
	"code-review-tg-bot/internal/vacation"
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StoryService struct {
	storage         storage.Storage
	stories         []Story
	vacationService *vacation.VacationService
	bot             *tg.BotAPI
	Users           *access.Users
	ReviewersIds    *access.UserIds
}

func NewStoryService(storage storage.Storage, stories []Story, vacationService *vacation.VacationService, bot *tg.BotAPI, users *access.Users, reviewersIds *access.UserIds) *StoryService {
	return &StoryService{
		storage:         storage,
		stories:         stories,
		vacationService: vacationService,
		bot:             bot,
		Users:           users,
		ReviewersIds:    reviewersIds,
	}
}

type StoryData struct {
	Name  string
	Stage string
	Extra interface{}
}

func (s *StoryService) getCurrentStoryData(userId int64) *StoryData {
	var storyData StoryData
	exists := s.storage.Get(fmt.Sprintf("active_story_%d", userId), &storyData)
	if !exists {
		return nil
	}
	return &storyData
}

func (s *StoryService) setCurrentStoryExtraData(userId int64, extra interface{}) {
	existingStoryData := s.getCurrentStoryData(userId)
	if existingStoryData != nil {
		existingStoryData.Extra = extra
		s.setCurrentStory(userId, existingStoryData)
	}
}

func (s *StoryService) getCurrentStoryExtraData(userId int64) interface{} {
	existingStoryData := s.getCurrentStoryData(userId)
	return existingStoryData.Extra
}

func (s *StoryService) setCurrentStory(userId int64, storyData *StoryData) {
	existingStoryData := s.getCurrentStoryData(userId)
	if existingStoryData != nil && existingStoryData.Extra != nil {
		storyData.Extra = existingStoryData.Extra
	}

	logger.Instance.Infow("setCurrentStory", "storyData", *storyData)
	s.storage.Set(fmt.Sprintf("active_story_%d", userId), *storyData)
}

func (s *StoryService) dropCurrentStory(userId int64) {
	s.storage.Delete(fmt.Sprintf("active_story_%d", userId))
}

func (s *StoryService) getStory(name string) *Story {
	for _, story := range s.stories {
		if story.name == name {
			return &story
		}
	}
	return nil
}

func (s *StoryService) HandleCurrentStories(update tg.Update, userId int64) bool {

	logger.Instance.Infow("HandleCurrentStories", "userId", userId)

	storyData := s.getCurrentStoryData(userId)
	logger.Instance.Infow("HandleCurrentStories", "storyData", storyData)
	if storyData == nil {
		return false
	}

	story := s.getStory(storyData.Name)
	if story == nil {
		return false
	}
	stage := story.getStage(storyData.Stage)
	if stage == nil {
		return false
	}

	logger.Instance.Infow("ExecuteStory", "Stage", stage.name)

	s.ExecuteStory(story, stage, update, userId)

	return true
}

func (s *StoryService) ExecuteStory(story *Story, stage *Stage, update tg.Update, userId int64) {

	logger.Instance.Infow("ExecuteStory", "userId", userId)

	logger.Instance.Infow("ExecuteStory", "story", story.name)

	if stage == nil {
		stage = &story.stages[0]
	}

	logger.Instance.Infow("ExecuteStory", "Stage", stage.name)

	executionStatus := stage.cb(update, s)
	logger.Instance.Infow("ExecuteStory", "executionStatus", executionStatus)
	if executionStatus == Dropped || executionStatus == Finished {
		s.dropCurrentStory(userId)
		return
	}
	if executionStatus == NotExecuted {
		s.setCurrentStory(userId, &StoryData{Name: story.name, Stage: stage.name})
		return
	}
	if executionStatus == Executed {
		nextStage := story.getNextStage(stage.name)

		if nextStage == nil {
			logger.Instance.Infow("ExecuteStory", "nextStage", nextStage)
			s.dropCurrentStory(userId)
		} else {
			logger.Instance.Infow("ExecuteStory", "nextStage", nextStage.name)
			s.setCurrentStory(userId, &StoryData{Name: story.name, Stage: nextStage.name})
		}
		return
	}
}
