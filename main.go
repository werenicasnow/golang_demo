package main

import (
	"fmt"
	"math"
)
	
const ImtPower = 2

func main() {
	fmt.Println("Калькулятор")
	userKg, userHeight := getUserInput()
	IMT := calculateIMT(userKg, userHeight)
	outputResult(IMT)
}

func generateShortCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:6]
}

func outputResult(imt  float64) {
	result := fmt.Sprintf("Ваш ИМТ: %.0f", imt)
	fmt.Print(result)
}

func calculateIMT(userKg float64, userHeight float64) (IMT float64) {
	IMT = userKg / math.Pow(userHeight / 100, ImtPower)
	return
}

func getUserInput() (float64, float64) {
	var userHeight float64
	var userKg float64

	fmt.Print("Введите рост в сантиметрах: ")
	fmt.Scan(&userHeight)
	fmt.Print("Введите вес в кг: ")
	fmt.Scan(&userKg)
	return userKg, userHeight
}