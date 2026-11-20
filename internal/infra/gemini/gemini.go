package gemini

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/mqqff/wasteco-be/internal/domain/env"
	"github.com/mqqff/wasteco-be/pkg/log"
	"google.golang.org/genai"
)

const (
	geminiTimeout       = 30 * time.Second
	defaultPrimaryModel = "gemini-flash-lite-latest"
)

var defaultFallbackModels = []string{
	"gemini-flash-latest",
	"gemini-2.5-flash",
	"gemini-2.5-flash-lite",
}

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
	AvailableIngredients  []RecipeIngredient `json:"available_ingredients"`
	AdditionalIngredients []RecipeIngredient `json:"additional_ingredients"`
	Steps                 []string           `json:"steps"`
	SmartTips             string             `json:"smart_tips"`
}

type RecipeResponse struct {
	Recipes []GeneratedRecipe `json:"recipes"`
}

type DetectedIngredient struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type detectResponse struct {
	Ingredients []DetectedIngredient `json:"ingredients"`
}

type ImageRecipeResponse struct {
	DetectedIngredients []DetectedIngredient `json:"detected_ingredients"`
	Recipes             []GeneratedRecipe    `json:"recipes"`
}

type GeminiItf interface {
	GenerateRecipes(ingredients []IngredientInput) ([]GeneratedRecipe, error)
	GenerateRecipesFromImage(image []byte, mimeType string, extraIngredients []IngredientInput) (ImageRecipeResponse, error)
	DetectIngredients(image []byte, mimeType string) ([]DetectedIngredient, error)
}

type generateFunc func(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)

type GeminiStruct struct {
	client     *genai.Client
	models     []string
	generateFn generateFunc
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

	return &GeminiStruct{
		client: client,
		models: modelCandidates(env.AppEnv.GeminiModel, env.AppEnv.GeminiFallbackModels),
	}
}

func modelCandidates(primary, fallbackEnv string) []string {
	primary = strings.TrimSpace(primary)
	if primary == "" {
		primary = defaultPrimaryModel
	}

	fallbacks := defaultFallbackModels
	if strings.TrimSpace(fallbackEnv) != "" {
		fallbacks = strings.Split(fallbackEnv, ",")
	}

	seen := map[string]bool{primary: true}
	candidates := []string{primary}
	for _, raw := range fallbacks {
		m := strings.TrimSpace(raw)
		if m == "" || seen[m] {
			continue
		}
		seen[m] = true
		candidates = append(candidates, m)
	}

	return candidates
}

func shouldFallback(err error) bool {
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var apiErr genai.APIError
	if errors.As(err, &apiErr) {
		return isRetryableStatusCode(apiErr.Code)
	}

	var apiErrPtr *genai.APIError
	if errors.As(err, &apiErrPtr) && apiErrPtr != nil {
		return isRetryableStatusCode(apiErrPtr.Code)
	}

	return true
}

func isRetryableStatusCode(code int) bool {
	switch code {
	case 404, 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

func (g *GeminiStruct) callModel(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	if g.generateFn != nil {
		return g.generateFn(ctx, model, contents, config)
	}
	if g.client == nil {
		return nil, fmt.Errorf("gemini client not initialized")
	}
	return g.client.Models.GenerateContent(ctx, model, contents, config)
}

func (g *GeminiStruct) generateWithFallback(caller string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	models := g.models
	if len(models) == 0 {
		models = modelCandidates(env.AppEnv.GeminiModel, env.AppEnv.GeminiFallbackModels)
	}

	var lastErr error
	for i, model := range models {
		ctx, cancel := context.WithTimeout(context.Background(), geminiTimeout)
		result, err := g.callModel(ctx, model, contents, config)
		cancel()

		if err == nil && result != nil && strings.TrimSpace(result.Text()) != "" {
			if i > 0 {
				log.Info(log.LogInfo{
					"model": model,
				}, "[Gemini]["+caller+"] succeeded using fallback model")
			}
			return result, nil
		}

		if err == nil {
			err = fmt.Errorf("empty response from model %s", model)
		}
		lastErr = err

		if !shouldFallback(err) || i == len(models)-1 {
			log.Error(log.LogInfo{
				"model": model,
				"error": err.Error(),
			}, "[Gemini]["+caller+"] failed to generate content from gemini")
			return nil, err
		}

		log.Info(log.LogInfo{
			"failed_model": model,
			"next_model":   models[i+1],
			"error":        err.Error(),
		}, "[Gemini]["+caller+"] model unavailable, falling back to next model")
	}

	return nil, lastErr
}

func recipeListSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeArray,
		Items: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"recipe_name":          {Type: genai.TypeString},
				"match_percentage":     {Type: genai.TypeInteger},
				"cooking_time_minutes": {Type: genai.TypeInteger},
				"difficulty":           {Type: genai.TypeString},
				"portions":             {Type: genai.TypeInteger},
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
				"difficulty", "portions",
				"available_ingredients", "additional_ingredients",
				"steps", "smart_tips",
			},
		},
	}
}

func ingredientListSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeArray,
		Items: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"name":     {Type: genai.TypeString},
				"quantity": {Type: genai.TypeNumber},
				"unit":     {Type: genai.TypeString},
			},
			Required: []string{"name", "quantity", "unit"},
		},
	}
}

func sortRecipesByMatch(recipes []GeneratedRecipe) []GeneratedRecipe {
	if recipes == nil {
		return []GeneratedRecipe{}
	}
	sort.SliceStable(recipes, func(i, j int) bool {
		return recipes[i].MatchPercentage > recipes[j].MatchPercentage
	})
	return recipes
}

func (g *GeminiStruct) GenerateRecipes(ingredients []IngredientInput) ([]GeneratedRecipe, error) {
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
4. 'steps' adalah instruksi memasak langkah demi langkah (antara 4 sampai 6 tahap) yang singkat dan jelas.
5. 'smart_tips' memberikan tips dapur praktis untuk menjaga cita rasa dan mencegah bahan tersisa lagi.
6. Seluruh teks wajib menggunakan Bahasa Indonesia yang ramah, jelas, dan menggugah selera.
`, sb.String())

	temp := float32(0.2)
	result, err := g.generateWithFallback(
		"GenerateRecipes",
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"recipes": recipeListSchema(),
				},
				Required: []string{"recipes"},
			},
			Temperature: &temp,
		},
	)
	if err != nil {
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

	return sortRecipesByMatch(recipeResp.Recipes), nil
}

func (g *GeminiStruct) GenerateRecipesFromImage(image []byte, mimeType string, extraIngredients []IngredientInput) (ImageRecipeResponse, error) {
	var extraSection strings.Builder
	if len(extraIngredients) > 0 {
		extraSection.WriteString("\nBahan tambahan yang juga tersedia dari pengguna:\n")
		for _, ing := range extraIngredients {
			extraSection.WriteString(fmt.Sprintf("- %s: %.2f %s\n", ing.Name, ing.Quantity, ing.Unit))
		}
	}

	prompt := fmt.Sprintf(`
Kamu adalah koki profesional dan asisten OCR spesialis Zero-Waste Food Management di Indonesia.
Tugasmu:
1. Kenali seluruh bahan makanan dari gambar yang dikirim, baik secara visual (bahan mentah, sayuran, buah, lauk, sisa makanan, isi kulkas) maupun melalui OCR teks (struk belanja, catatan daftar bahan, label kemasan makanan).
2. Masukkan hasil pengenalan bahan dari gambar ke 'detected_ingredients' ('name' dalam Bahasa Indonesia, perkiraan 'quantity' angka, dan 'unit' satuan rumah tangga yang wajar). Abaikan benda atau teks non-makanan seperti harga, nama toko, atau tanggal struk.
3. Berdasarkan bahan yang terdeteksi dari gambar%s, berikan 2 sampai 3 rekomendasi resep masakan rumahan Indonesia yang lezat, praktis, dan zero-waste pada 'recipes'.
4. Jika gambar sama sekali tidak mengandung bahan makanan maupun teks bahan makanan, kembalikan 'detected_ingredients' kosong dan 'recipes' kosong.

Aturan resep:
1. Prioritaskan penggunaan bahan yang terdeteksi sebanyak mungkin.
2. Tambahkan hanya bumbu/bahan dasar umum rumah tangga Indonesia (minyak, garam, bawang putih/merah, kecap, merica) pada 'additional_ingredients'.
3. 'match_percentage' bernilai 50 sampai 95 dan diurutkan dari tertinggi ke terendah.
4. 'steps' berisi 4 sampai 6 tahap memasak singkat dan jelas dalam Bahasa Indonesia.
5. 'smart_tips' memberikan tips praktis menjaga cita rasa dan mencegah bahan tersisa lagi.
`, extraSection.String())

	temp := float32(0.2)
	contents := []*genai.Content{
		genai.NewContentFromParts([]*genai.Part{
			genai.NewPartFromBytes(image, mimeType),
			genai.NewPartFromText(prompt),
		}, genai.RoleUser),
	}

	result, err := g.generateWithFallback(
		"GenerateRecipesFromImage",
		contents,
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"detected_ingredients": ingredientListSchema(),
					"recipes":              recipeListSchema(),
				},
				Required: []string{"detected_ingredients", "recipes"},
			},
			Temperature: &temp,
		},
	)
	if err != nil {
		return ImageRecipeResponse{}, err
	}

	var resp ImageRecipeResponse
	if err := sonic.Unmarshal([]byte(result.Text()), &resp); err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
			"raw":   result.Text(),
		}, "[Gemini][GenerateRecipesFromImage] failed to unmarshal json")
		return ImageRecipeResponse{}, err
	}

	if resp.DetectedIngredients == nil {
		resp.DetectedIngredients = []DetectedIngredient{}
	}
	resp.Recipes = sortRecipesByMatch(resp.Recipes)

	return resp, nil
}

func (g *GeminiStruct) DetectIngredients(image []byte, mimeType string) ([]DetectedIngredient, error) {
	prompt := `
Kamu adalah asisten dapur dengan kemampuan OCR dan pengenalan visual untuk aplikasi pengelolaan bahan sisa di Indonesia.
Identifikasi setiap bahan makanan yang terlihat di foto maupun dari teks OCR (seperti struk belanja, label kemasan, atau catatan bahan).

Aturan penting:
1. 'name' adalah nama bahan dalam Bahasa Indonesia (contoh: Bayam, Tomat, Bawang Putih, Nasi Putih).
2. 'quantity' adalah perkiraan jumlah yang terlihat atau tertulis, 'unit' adalah satuan rumah tangga yang wajar (contoh: ikat, buah, siung, butir, porsi, gram, lembar).
3. Abaikan benda atau teks non-makanan (seperti harga atau nama toko). Jika tidak ada bahan makanan, kembalikan daftar kosong.
`

	temp := float32(0.1)
	contents := []*genai.Content{
		genai.NewContentFromParts([]*genai.Part{
			genai.NewPartFromBytes(image, mimeType),
			genai.NewPartFromText(prompt),
		}, genai.RoleUser),
	}

	result, err := g.generateWithFallback(
		"DetectIngredients",
		contents,
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"ingredients": ingredientListSchema(),
				},
				Required: []string{"ingredients"},
			},
			Temperature: &temp,
		},
	)
	if err != nil {
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

	if resp.Ingredients == nil {
		resp.Ingredients = []DetectedIngredient{}
	}

	return resp.Ingredients, nil
}
