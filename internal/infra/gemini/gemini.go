package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/mqqff/wasteco-be/internal/domain/env"
	"github.com/mqqff/wasteco-be/pkg/log"
	"google.golang.org/genai"
)

type IngredientInput struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type RecipeIngredient struct {
	Name     string `json:"name"`
	Quantity string `json:"quantity"`
}

type GeneratedRecipe struct {
	RecipeName            string             `json:"recipe_name"`
	MatchPercentage       int                `json:"match_percentage"`
	CookingTimeMinutes    int                `json:"cooking_time_minutes"`
	Difficulty            string             `json:"difficulty"`
	Portions              int                `json:"portions"`
	EstimatedSavings      float64            `json:"estimated_savings"`
	AvailableIngredients  []RecipeIngredient `json:"available_ingredients"`
	AdditionalIngredients []RecipeIngredient `json:"additional_ingredients"`
	Steps                 []string           `json:"steps"`
	SmartTips             string             `json:"smart_tips"`
}

type RecipeResponse struct {
	Recipes []GeneratedRecipe `json:"recipes"`
}

type DetectedIngredient struct {
	Name      string  `json:"name"`
	Quantity  float64 `json:"quantity"`
	Unit      string  `json:"unit"`
	UnitPrice float64 `json:"unit_price"`
}

type detectResponse struct {
	Ingredients []DetectedIngredient `json:"ingredients"`
}

type GeminiItf interface {
	GenerateRecipes(ingredients []IngredientInput) ([]GeneratedRecipe, error)
	DetectIngredients(image []byte, mimeType string) ([]DetectedIngredient, error)
}

type GeminiStruct struct {
	client *genai.Client
}

var Gemini = getGemini()

func getGemini() GeminiItf {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: env.AppEnv.GeminiAPIKey,
	})
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[Gemini][getGemini] failed to create gemini client")
	}

	return &GeminiStruct{client: client}
}

func (g *GeminiStruct) GenerateRecipes(ingredients []IngredientInput) ([]GeneratedRecipe, error) {
	if g.client == nil {
		return nil, fmt.Errorf("gemini client not initialized")
	}

	var sb strings.Builder
	sb.WriteString("Daftar bahan makanan dapur yang tersedia:\n")
	for _, ing := range ingredients {
		sb.WriteString(fmt.Sprintf("- %s: %.2f %s\n", ing.Name, ing.Quantity, ing.Unit))
	}

	prompt := fmt.Sprintf(`
Kamu adalah koki profesional spesialis Zero-Waste Food Management di Indonesia.
Tugasmu adalah memberikan 2 sampai 3 rekomendasi resep masakan rumahan lezat, praktis, dan hemat menggunakan bahan-bahan dapur yang diberikan di bawah ini.

%s

Aturan penting:
1. Prioritaskan penggunaan bahan yang tersedia sebanyak mungkin.
2. Tambahkan hanya bumbu/bahan dasar umum rumah tangga Indonesia (seperti minyak, garam, bawang putih/merah, kecap, merica) pada 'additional_ingredients'.
3. 'match_percentage' adalah persentase kecocokan bahan dapur yang tersedia (nilai 50 sampai 95). Urutkan resep dari kecocokan tertinggi ke terendah.
4. 'estimated_savings' adalah estimasi perkiraan rupiah yang dihemat dengan mengolah bahan sisa ini dibanding makan di luar atau membeli bahan baru (contoh: 12000, 15000, 20000).
5. 'steps' adalah instruksi memasak langkah demi langkah (antara 4 sampai 6 tahap) yang singkat dan jelas.
6. 'smart_tips' memberikan tips dapur praktis untuk menjaga cita rasa dan mencegah bahan tersisa lagi.
7. Seluruh teks wajib menggunakan Bahasa Indonesia yang ramah, jelas, dan menggugah selera.
`, sb.String())

	temp := float32(0.2)
	modelName := env.AppEnv.GeminiModel
	if modelName == "" {
		modelName = "gemini-flash-lite-latest"
	}

	result, err := g.client.Models.GenerateContent(
		context.Background(),
		modelName,
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"recipes": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"recipe_name":          {Type: genai.TypeString},
								"match_percentage":     {Type: genai.TypeInteger},
								"cooking_time_minutes": {Type: genai.TypeInteger},
								"difficulty":           {Type: genai.TypeString},
								"portions":             {Type: genai.TypeInteger},
								"estimated_savings":    {Type: genai.TypeNumber},
								"available_ingredients": {
									Type: genai.TypeArray,
									Items: &genai.Schema{
										Type: genai.TypeObject,
										Properties: map[string]*genai.Schema{
											"name":     {Type: genai.TypeString},
											"quantity": {Type: genai.TypeString},
										},
										Required: []string{"name", "quantity"},
									},
								},
								"additional_ingredients": {
									Type: genai.TypeArray,
									Items: &genai.Schema{
										Type: genai.TypeObject,
										Properties: map[string]*genai.Schema{
											"name":     {Type: genai.TypeString},
											"quantity": {Type: genai.TypeString},
										},
										Required: []string{"name", "quantity"},
									},
								},
								"steps": {
									Type:  genai.TypeArray,
									Items: &genai.Schema{Type: genai.TypeString},
								},
								"smart_tips": {Type: genai.TypeString},
							},
							Required: []string{
								"recipe_name", "match_percentage", "cooking_time_minutes",
								"difficulty", "portions", "estimated_savings",
								"available_ingredients", "additional_ingredients",
								"steps", "smart_tips",
							},
						},
					},
				},
				Required: []string{"recipes"},
			},
			Temperature: &temp,
		},
	)

	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[Gemini][GenerateRecipes] failed to generate content from gemini")
		return nil, err
	}

	var recipeResp RecipeResponse
	err = sonic.Unmarshal([]byte(result.Text()), &recipeResp)
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
			"raw":   result.Text(),
		}, "[Gemini][GenerateRecipes] failed to unmarshal json")
		return nil, err
	}

	return recipeResp.Recipes, nil
}

func (g *GeminiStruct) DetectIngredients(image []byte, mimeType string) ([]DetectedIngredient, error) {
	if g.client == nil {
		return nil, fmt.Errorf("gemini client not initialized")
	}

	prompt := `
Kamu adalah asisten dapur yang mengenali bahan makanan dari foto untuk aplikasi pengelolaan bahan sisa di Indonesia.
Identifikasi setiap bahan makanan mentah atau sisa yang terlihat di foto.

Aturan penting:
1. 'name' adalah nama bahan dalam Bahasa Indonesia (contoh: Bayam, Tomat, Bawang Putih, Nasi Putih).
2. 'quantity' adalah perkiraan jumlah yang terlihat, 'unit' adalah satuan rumah tangga yang wajar (contoh: ikat, buah, siung, butir, porsi, gram, lembar).
3. 'unit_price' adalah perkiraan harga pasar per SATU satuan dalam Rupiah (contoh: bayam 3000 per ikat, tomat 1000 per buah, wortel 30 per gram).
4. Abaikan benda yang bukan bahan makanan. Jika tidak ada bahan makanan, kembalikan daftar kosong.
`

	temp := float32(0.1)
	modelName := env.AppEnv.GeminiModel
	if modelName == "" {
		modelName = "gemini-flash-lite-latest"
	}

	contents := []*genai.Content{
		genai.NewContentFromParts([]*genai.Part{
			genai.NewPartFromBytes(image, mimeType),
			genai.NewPartFromText(prompt),
		}, genai.RoleUser),
	}

	result, err := g.client.Models.GenerateContent(
		context.Background(),
		modelName,
		contents,
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"ingredients": {
						Type: genai.TypeArray,
						Items: &genai.Schema{
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"name":       {Type: genai.TypeString},
								"quantity":   {Type: genai.TypeNumber},
								"unit":       {Type: genai.TypeString},
								"unit_price": {Type: genai.TypeNumber},
							},
							Required: []string{"name", "quantity", "unit", "unit_price"},
						},
					},
				},
				Required: []string{"ingredients"},
			},
			Temperature: &temp,
		},
	)
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[Gemini][DetectIngredients] failed to generate content from gemini")
		return nil, err
	}

	var resp detectResponse
	if err := sonic.Unmarshal([]byte(result.Text()), &resp); err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
			"raw":   result.Text(),
		}, "[Gemini][DetectIngredients] failed to unmarshal json")
		return nil, err
	}

	return resp.Ingredients, nil
}
