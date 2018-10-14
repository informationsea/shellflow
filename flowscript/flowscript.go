package flowscript

func ParseScript(text string) (Evaluable, error) {
	tokenizer := NewTokenizerFromText(text)
	tokenizer.Scan()
	return ParseAsExp(tokenizer)
}

func EvaluateScript(text string, env Environment) (Value, error) {
	eval, err := ParseScript(text)
	if err != nil {
		return nil, err
	}
	return eval.Evaluate(env)
}

func SearchDependentVariables(evaluable Evaluable) StringSet {
	variables := NewStringSet()

	if ae, ok := evaluable.(*AssignExpression); ok {
		evaluable = ae.exp
	}

	if ae, ok := evaluable.(*Variable); ok {
		variables.Add(ae.Name)
	} else {
		for _, v := range evaluable.SubEvaluable() {
			variables.AddAll(SearchDependentVariables(v))
		}
	}

	return variables
}

func SearchCreatedVariables(evaluable Evaluable) StringSet {
	variables := NewStringSet()

	if ae, ok := evaluable.(*AssignExpression); ok {
		if ve, ok := ae.variable.(*Variable); ok {
			variables.Add(ve.Name)
		}
	} else {
		for _, v := range evaluable.SubEvaluable() {
			variables.AddAll(SearchCreatedVariables(v))
		}
	}

	return variables
}
