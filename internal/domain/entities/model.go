package entities

// dependendo de um escopo um Model é imutável
// o struct é um modelo parecido com uma classe em orientação objeto
type Model struct {
	Name      string
	MaxTokens int
}

// função construtora | * representa um ponteiro que referencia a Model
func NewModel(name string, maxTokens int) *Model {
	return &Model{
		Name:      name,
		MaxTokens: maxTokens,
	}
}

func (m *Model) GetMaxTokens() int {
	return m.MaxTokens
}

func (m *Model) GetModelName() string {
	return m.Name
}
