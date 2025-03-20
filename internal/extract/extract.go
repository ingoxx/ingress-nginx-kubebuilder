package extract

type ExtractIng interface {
	GetIngressAnnotations()
	GetInsData()
}

type Extract struct {
}

func (e *Extract) GetIngressAnnotations() {}

func (e *Extract) GetInsData() {

}
