module github.com/jlfowle/sonarqube-operator

go 1.13

require (
	github.com/magiconair/properties v1.8.0
	github.com/operator-framework/operator-sdk v0.18.2
	github.com/parflesh/sonarqube-operator v0.0.0-20200608154349-0979d6cffde2 // indirect
	github.com/spf13/pflag v1.0.5
	golang.org/x/mod v0.2.0
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.0
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/client-go => k8s.io/client-go v0.18.2 // Required by prometheus-operator
)
