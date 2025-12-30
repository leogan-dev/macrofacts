package auth

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type MeResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type MeSettingsResponse struct {
	Timezone     string `json:"timezone"`
	CalorieGoal  int    `json:"calorieGoal"`
	ProteinGoalG int    `json:"proteinGoalG"`
	CarbsGoalG   int    `json:"carbsGoalG"`
	FatGoalG     int    `json:"fatGoalG"`
}

type UpdateSettingsRequest struct {
	Timezone     *string `json:"timezone,omitempty"`
	CalorieGoal  *int    `json:"calorieGoal,omitempty"`
	ProteinGoalG *int    `json:"proteinGoalG,omitempty"`
	CarbsGoalG   *int    `json:"carbsGoalG,omitempty"`
	FatGoalG     *int    `json:"fatGoalG,omitempty"`
}
