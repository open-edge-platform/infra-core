// SPDX-FileCopyrightText: (C) 2025 Intel Corporation
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	mathrand "math/rand"
)

func RandomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, length)
	for i := range b {
		//nolint:gosec // math/rand is used for testing only
		b[i] = letters[mathrand.Intn(len(letters))]
	}
	return string(b)
}

func GenerateRandomSha256() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(GenerateRandomProfileName())))
}

func GenerateRandomProfileName() string {
	return fmt.Sprintf("Test OS profile name #%d", generateRandomInteger(1023))
}

func GenerateRandomOsResourceName() string {
	return fmt.Sprintf("Test OS name #%d", generateRandomInteger(1023))
}

func GenerateRandomOsUpdatePolicyName() string {
	return fmt.Sprintf("Test OS Update Policy name #%d", generateRandomInteger(1023))
}

func GenerateRandInt(minValue, maxValue int) int64 {
	nBig, err := rand.Int(rand.Reader, new(big.Int).SetUint64(uint64(maxValue-minValue+1))) //nolint:gosec // Testing only
	if err != nil {
		panic(err)
	}

	return nBig.Int64() + int64(minValue)
}

func generateRandomInteger(intMax int64) int64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(intMax))
	if err != nil {
		panic(err)
	}
	n := nBig.Int64()
	return n
}
