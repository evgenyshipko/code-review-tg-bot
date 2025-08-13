package stories

import (
	"errors"
)

func getPickedUserIdFromExtraData(s *StoryService, userId int64) (int64, error) {
	pickedUserIdRaw := s.getCurrentStoryExtraData(userId)
	pickedUserIdFloat, ok := pickedUserIdRaw.(float64)
	if !ok {
		return 0, errors.New("не можем привести id к float64")
	}

	pickedUserId := int64(pickedUserIdFloat)
	return pickedUserId, nil
}
