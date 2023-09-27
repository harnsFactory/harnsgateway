package runtime

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type ValidateNameFunc func(name string) error

func Validate(name string, nameFn ValidateNameFunc) field.ErrorList {
	return ValidateObjectMeta(name, nameFn)
}

func ValidateObjectMeta(name string, nameFn ValidateNameFunc) field.ErrorList {
	var allErrs field.ErrorList
	if len(name) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("name"), ""))
	} else if err := nameFn(name); err != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("name"), name, err.Error()))
	}
	return allErrs
}
