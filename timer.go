package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

func startTimer(timerInMinutes string, configuration Configuration) {
	timeoutInMinutes := toMinutes(timerInMinutes)

	timeoutInSeconds := timeoutInMinutes * 60
	timeOfTimeout := time.Now().Add(time.Minute * time.Duration(timeoutInMinutes)).Format("15:04")
	debugInfo(fmt.Sprintf("Starting timer at %s for %d minutes = %d seconds (parsed from user input %s)", timeOfTimeout, timeoutInMinutes, timeoutInSeconds, timerInMinutes))

	timerSuccessful := false

	room := getMobTimerRoom(configuration)
	if room != "" {
		timerUser := getUserForMobTimer(configuration.TimerUser)
		err := httpPutTimer(timeoutInMinutes, room, timerUser, configuration.TimerUrl)
		if err != nil {
			sayError("remote timer couldn't be started")
			sayError(err.Error())
		} else {
			timerSuccessful = true
		}
	}

	if configuration.TimerLocal {
		err := executeCommandsInBackgroundProcess(getSleepCommand(timeoutInSeconds), getVoiceCommand(configuration.VoiceMessage, configuration.VoiceCommand), getNotifyCommand(configuration.NotifyMessage, configuration.NotifyCommand))

		if err != nil {
			sayError(fmt.Sprintf("timer couldn't be started on your system (%s)", runtime.GOOS))
			sayError(err.Error())
		} else {
			timerSuccessful = true
		}
	}

	if timerSuccessful {
		sayInfo("It's now " + currentTime() + ". " + fmt.Sprintf("%d min timer ends at approx. %s", timeoutInMinutes, timeOfTimeout) + ". Happy collaborating! :)")
	}
}

func getMobTimerRoom(configuration Configuration) string {
	currentWipBranchQualifier := configuration.WipBranchQualifier
	if currentWipBranchQualifier == "" {
		currentBranch := gitCurrentBranch()
		currentBaseBranch, _ := determineBranches(currentBranch, gitBranches(), configuration)

		if currentBranch.IsWipBranch(configuration) {
			wipBranchWithoutWipPrefix := currentBranch.removeWipPrefix(configuration).Name
			currentWipBranchQualifier = removePrefix(removePrefix(wipBranchWithoutWipPrefix, currentBaseBranch.Name), configuration.WipBranchQualifierSeparator)
		}
	}

	if configuration.TimerRoomUseWipBranchQualifier && currentWipBranchQualifier != "" {
		sayInfo("Using wip branch qualifier for room name")
		return currentWipBranchQualifier
	}
	return configuration.TimerRoom
}

func startBreakTimer(timerInMinutes string, configuration Configuration) {
	timeoutInMinutes := toMinutes(timerInMinutes)

	timeoutInSeconds := timeoutInMinutes * 60
	timeOfTimeout := time.Now().Add(time.Minute * time.Duration(timeoutInMinutes)).Format("15:04")
	debugInfo(fmt.Sprintf("Starting break timer at %s for %d minutes = %d seconds (parsed from user input %s)", timeOfTimeout, timeoutInMinutes, timeoutInSeconds, timerInMinutes))

	timerSuccessful := false
	room := getMobTimerRoom(configuration)
	if room != "" {
		timerUser := getUserForMobTimer(configuration.TimerUser)
		err := httpPutBreakTimer(timeoutInMinutes, room, timerUser, configuration.TimerUrl)
		if err != nil {
			sayError("remote break timer couldn't be started")
			sayError(err.Error())
		} else {
			timerSuccessful = true
		}
	}

	if configuration.TimerLocal {
		err := executeCommandsInBackgroundProcess(getSleepCommand(timeoutInSeconds), getVoiceCommand("mob start", configuration.VoiceCommand), getNotifyCommand("mob start", configuration.NotifyCommand))

		if err != nil {
			sayError(fmt.Sprintf("break timer couldn't be started on your system (%s)", runtime.GOOS))
			sayError(err.Error())
		} else {
			timerSuccessful = true
		}
	}

	if timerSuccessful {
		sayInfo("It's now " + currentTime() + ". " + fmt.Sprintf("%d min break timer ends at approx. %s", timeoutInMinutes, timeOfTimeout) + ". Happy collaborating! :)")
	}
}

func getUserForMobTimer(userOverride string) string {
	if userOverride == "" {
		return gitUserName()
	}
	return userOverride
}

func toMinutes(timerInMinutes string) int {
	timeoutInMinutes, _ := strconv.Atoi(timerInMinutes)
	if timeoutInMinutes < 0 {
		timeoutInMinutes = 0
	}
	return timeoutInMinutes
}

func httpPutTimer(timeoutInMinutes int, room string, user string, timerService string) error {
	putBody, _ := json.Marshal(map[string]interface{}{
		"timer": timeoutInMinutes,
		"user":  user,
	})
	return sendRequest(putBody, "PUT", timerService+room)
}

func httpPutBreakTimer(timeoutInMinutes int, room string, user string, timerService string) error {
	putBody, _ := json.Marshal(map[string]interface{}{
		"breaktimer": timeoutInMinutes,
		"user":       user,
	})
	return sendRequest(putBody, "PUT", timerService+room)
}

func sendRequest(requestBody []byte, requestMethod string, requestUrl string) error {
	sayInfo(requestMethod + " " + requestUrl + " " + string(requestBody))

	responseBody := bytes.NewBuffer(requestBody)
	request, requestCreationError := http.NewRequest(requestMethod, requestUrl, responseBody)
	if requestCreationError != nil {
		return fmt.Errorf("failed to create the http request object: %w", requestCreationError)
	}

	request.Header.Set("Content-Type", "application/json")
	response, responseErr := http.DefaultClient.Do(request)
	if responseErr != nil {
		return fmt.Errorf("failed to make the http request: %w", responseErr)
	}
	defer response.Body.Close()
	body, responseReadingErr := ioutil.ReadAll(response.Body)
	if responseReadingErr != nil {
		return fmt.Errorf("failed to read the http response: %w", responseReadingErr)
	}
	if string(body) != "" {
		sayInfo(string(body))
	}
	return nil
}

func currentTime() string {
	return time.Now().Format("15:04")
}
