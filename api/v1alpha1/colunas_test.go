package v1alpha1

import (
	"os"
	"testing"

	"sigs.k8s.io/yaml"
)

// As colunas de `kubectl get` não têm teste de compilação: a marcação é
// comentário, o CRD é gerado, e uma coluna que imprime lixo continua passando
// em tudo. Este teste lê o CRD gerado, que é o que o cluster recebe.
type crdDeTeste struct {
	Spec struct {
		Versions []struct {
			Name                     string `json:"name"`
			AdditionalPrinterColumns []struct {
				Name     string `json:"name"`
				Type     string `json:"type"`
				JSONPath string `json:"jsonPath"`
			} `json:"additionalPrinterColumns"`
		} `json:"versions"`
	} `json:"spec"`
}

func colunas(t *testing.T) map[string]string {
	t.Helper()

	cru, err := os.ReadFile("../../config/crd/bases/preview.rigo.dev_previewenvironments.yaml")
	if err != nil {
		t.Fatalf("ler o CRD gerado: %v", err)
	}

	var crd crdDeTeste
	if err := yaml.Unmarshal(cru, &crd); err != nil {
		t.Fatalf("decodificar o CRD: %v", err)
	}

	tipos := map[string]string{}
	for _, versao := range crd.Spec.Versions {
		for _, coluna := range versao.AdditionalPrinterColumns {
			tipos[coluna.Name] = coluna.Type
		}
	}
	if len(tipos) == 0 {
		t.Fatal("o CRD gerado não tem nenhuma printer column")
	}
	return tipos
}

// O kubectl imprime coluna `type: date` como tempo decorrido. Para um instante
// no futuro a subtração dá negativo e a saída vira `<invalid>` — a coluna do
// vencimento ficava ilegível durante toda a vida do ambiente, que é o único
// momento em que alguém olha para ela. Só depois de o ambiente já ter vencido
// é que ela passava a mostrar alguma coisa.
func TestColunaDeVencimentoNaoEDate(t *testing.T) {
	tipos := colunas(t)

	tipo, existe := tipos["Expira"]
	if !existe {
		t.Fatal("a coluna Expira sumiu do CRD")
	}
	if tipo == "date" {
		t.Error("Expira voltou a ser type=date; o kubectl vai imprimir <invalid> " +
			"enquanto o ambiente estiver de pé, porque o vencimento está no futuro")
	}
}

// A contrapartida: Idade aponta para creationTimestamp, que é sempre passado.
// Ali `date` é o tipo certo, e trocá-lo por string encheria a coluna de
// timestamp em vez do "2m" que se quer ler de relance.
func TestColunaDeIdadeContinuaDate(t *testing.T) {
	tipos := colunas(t)

	if tipos["Idade"] != "date" {
		t.Errorf("Idade deveria ser type=date, é %q", tipos["Idade"])
	}
}
