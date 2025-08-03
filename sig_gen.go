package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	unsafe_rand "math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

const (
	MAX_REQ_TIME      = 500 * time.Millisecond
	ROTATE_GRACE_TIME = 30 * time.Second
	letterBytes       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits     = 6
	letterIdxMask     = 1<<letterIdxBits - 1
	letterIdxMax      = 63 / letterIdxBits
)

type StaticSignature struct {
	CurrentSignature  string
	PreviousSignature string
	Logger            *zerolog.Logger
}

func (signature *StaticSignature) DoSignaturesMatch(compare string) bool {
	return (compare == signature.CurrentSignature || (signature.PreviousSignature != "" && compare == signature.PreviousSignature))
}

func (signature *StaticSignature) startSignatureGeneration() {
	go signature.updateSignature()
}

func (signature *StaticSignature) updateSignature() {
	for {
		if signature.CurrentSignature != "" {
			signature.PreviousSignature = signature.CurrentSignature
		}
		signature.CurrentSignature = signature.GenerateSignature()
		signature.Logger.Info().Dur("graceTime", ROTATE_GRACE_TIME).Any("signature", signature).Msg("Signature rotating.")
		time.Sleep(ROTATE_GRACE_TIME)
		signature.PreviousSignature = ""
		signature.Logger.Info().Any("signature", signature).Msg("Previous signature removed.")
		signature.awaitRandomTime()
	}
}

func (signature *StaticSignature) awaitRandomTime() {
	minimumWaitMinute := 10
	maximumWaitMinute := (60 + 1) - minimumWaitMinute
	src := rand.Reader
	randNum, err := rand.Int(src, big.NewInt(int64(maximumWaitMinute)))
	if err != nil {
		signature.Logger.Error().Msg("Unable to generate in signature generation timeout... Defaulting to unsafe random!")
		unsafeRandInt := unsafe_rand.Int31n(int32(maximumWaitMinute))
		duration := time.Duration(int32(minimumWaitMinute) + unsafeRandInt)
		time.Sleep(duration * time.Minute)
	} else {
		waitMinutes := minimumWaitMinute + int(randNum.Int64())
		signature.Logger.Info().Int("waitMinutes", waitMinutes).Msg("New signature generation queued.")
		duration := time.Duration(waitMinutes)
		time.Sleep(duration * time.Minute)
	}
}

func (signature *StaticSignature) RenderWithDuration(c *fiber.Ctx, fileName string, data fiber.Map, layouts ...string) error {
	if data == nil {
		data = fiber.Map{}
	}

	end := time.Since((c.Locals("startTime").(time.Time)))

	if end > MAX_REQ_TIME {
		logContext(signature.Logger.Warn(), c).
			Str("elapsed", end.String()).
			Str("uri", c.OriginalURL()).
			Msg("Server took a long time to respond.")
	}

	data["StaticSignature"] = signature.CurrentSignature

	return c.Render(fileName, data, layouts...)
}

func (signature *StaticSignature) GenerateSignature() string {
	hmacObj := hmac.New(sha256.New, GenerateSecret())
	hmacObj.Write(GenerateSecret())
	return hex.EncodeToString(hmacObj.Sum(nil))
}

func GenerateSecret() []byte {
	n := 32
	b := make([]byte, n)

	src := rand.Reader
	randNum, err := rand.Int(src, big.NewInt(63))
	if err != nil {
		WEB_LOGGER.Fatal().Msg("Unable to generate initial random int.")
	}

	for i, cache, remain := n-1, randNum.Int64(), letterIdxMax; i >= 0; {
		if remain == 0 {
			newRandNum, err := rand.Int(src, big.NewInt(63))
			if err != nil {
				WEB_LOGGER.Fatal().Msg("Unable to generate a new random number.")
			}
			cache, remain = newRandNum.Int64(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return b
}
