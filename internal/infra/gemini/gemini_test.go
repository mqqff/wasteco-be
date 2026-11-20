package gemini

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"google.golang.org/genai"
)

func TestMain(m *testing.M) {
	if root, err := filepath.Abs("../../.."); err == nil {
		_ = os.Chdir(root)
	}
	os.Exit(m.Run())
}

func TestModelCandidatesDeduplication(t *testing.T) {
	got := modelCandidates("gemini-flash-latest", "gemini-flash-latest, gemini-2.5-flash, , gemini-2.5-flash")
	want := []string{"gemini-flash-latest", "gemini-2.5-flash"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("modelCandidates = %v, want %v", got, want)
	}
}

func TestFallbackOnUnavailableModelAndStopOnAuthError(t *testing.T) {
	var called []string
	g := &GeminiStruct{
		models: []string{"primary-model", "fallback-1", "fallback-2"},
		generateFn: func(_ context.Context, model string, _ []*genai.Content, _ *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
			called = append(called, model)
			if model == "primary-model" {
				return nil, genai.APIError{Code: 503, Status: "UNAVAILABLE", Message: "overloaded"}
			}
			return &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []*genai.Part{
								{Text: `{"recipes":[{"recipe_name":"Nasi Goreng","match_percentage":90,"cooking_time_minutes":10,"difficulty":"Mudah","portions":2,"available_ingredients":[],"additional_ingredients":[],"steps":["Goreng"],"smart_tips":"Tips"}]}`},
							},
						},
					},
				},
			}, nil
		},
	}

	recipes, err := g.GenerateRecipes([]IngredientInput{{Name: "Nasi", Quantity: 1, Unit: "porsi"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(recipes) != 1 || recipes[0].RecipeName != "Nasi Goreng" {
		t.Fatalf("unexpected recipes: %+v", recipes)
	}
	if !reflect.DeepEqual(called, []string{"primary-model", "fallback-1"}) {
		t.Fatalf("called models = %v, want [primary-model fallback-1]", called)
	}

	called = nil
	g.generateFn = func(_ context.Context, model string, _ []*genai.Content, _ *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
		called = append(called, model)
		return nil, genai.APIError{Code: 403, Status: "PERMISSION_DENIED", Message: "invalid key"}
	}
	if _, err := g.DetectIngredients([]byte("img"), "image/jpeg"); err == nil {
		t.Fatal("expected error on 403")
	}
	if !reflect.DeepEqual(called, []string{"primary-model"}) {
		t.Fatalf("403 should not retry fallback, called = %v", called)
	}
}

func TestGenerateRecipesFromImageSortsAndParsesOCR(t *testing.T) {
	g := &GeminiStruct{
		models: []string{"primary-model"},
		generateFn: func(_ context.Context, _ string, _ []*genai.Content, _ *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
			return &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []*genai.Part{
								{Text: `{"detected_ingredients":[{"name":"Bayam","quantity":1,"unit":"ikat"}],"recipes":[{"recipe_name":"Tumis","match_percentage":70,"cooking_time_minutes":10,"difficulty":"Mudah","portions":1,"available_ingredients":[],"additional_ingredients":[],"steps":["Tumis"],"smart_tips":"Tips"},{"recipe_name":"Sup Bayam","match_percentage":95,"cooking_time_minutes":15,"difficulty":"Mudah","portions":2,"available_ingredients":[],"additional_ingredients":[],"steps":["Rebus"],"smart_tips":"Tips"}]}`},
							},
						},
					},
				},
			}, nil
		},
	}

	res, err := g.GenerateRecipesFromImage([]byte("img"), "image/jpeg", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.DetectedIngredients) != 1 || res.DetectedIngredients[0].Name != "Bayam" {
		t.Fatalf("unexpected detected ingredients: %+v", res.DetectedIngredients)
	}
	if len(res.Recipes) != 2 || res.Recipes[0].RecipeName != "Sup Bayam" {
		t.Fatalf("recipes not sorted by match_percentage: %+v", res.Recipes)
	}
}
