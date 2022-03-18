package types

type KCOConfig struct {
	TargetNamespaces  []string `yaml:"targetNamespaces"`
	ResourcesToIgnore struct {
		Deployments []struct {
			Namespace string `yaml:"namespace"`
			Name      string `yaml:"name"`
		} `yaml:"deployments"`
		StatefulSets []struct {
			Namespace string `yaml:"namespace"`
			Name      string `yaml:"name"`
		} `yaml:"statefuleSets"`
	} `yaml:"resourcesToIgnore"`
	// Required
	SpecPatch struct {
		Tolerations []struct {
			Key      string `yaml:"key"`
			Operator string `yaml:"operator"`
			Value    string `yaml:"value"`
			Effect   string `yaml:"effect"`
			// TolerationSeconds *int64 `yaml:"tolerationSeconds"`
		} `yaml:"tolerations"`
		Affinity struct {
			NodeAffinity struct {
				PreferredDuringSchedulingIgnoredDuringExecution []struct {
					Weight     int32 `yaml:"weight"`
					Preference struct {
						MatchExpressions []struct {
							Key      string   `yaml:"key"`
							Operator string   `yaml:"operator"`
							Values   []string `yaml:"values"`
						} `yaml:"matchExpressions"`
					} `yaml:"preference"`
				} `yaml:"preferredDuringSchedulingIgnoredDuringExecution"`
			} `yaml:"nodeAffinity"`
		} `yaml:"affinity"`
	} `yaml:"specPatch"`
}
