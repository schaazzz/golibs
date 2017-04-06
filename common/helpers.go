package common

func CheckAgainst(input string, validValues ...string) bool {
    for _, element := range validValues {
        if input == element {
            return true
        }
    }

    return false
}


