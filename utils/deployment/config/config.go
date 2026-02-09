package config

import (
	"fmt"
	"reflect"
)

// PopulateConfig 将配置映射到目标结构体
// config: 配置数据
// target: 目标结构体指针
func PopulateConfig(config map[string]interface{}, target interface{}) error {
	targetVal := reflect.ValueOf(target)
	if targetVal.Kind() != reflect.Ptr || targetVal.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}

	targetType := targetVal.Elem().Type()
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldVal := targetVal.Elem().Field(i)

		// 获取配置键名（支持json标签）
		key := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			key = jsonTag
		}

		// 检查配置中是否存在该键
		if value, ok := config[key]; ok {
			// 尝试设置字段值
			if err := setFieldValue(fieldVal, value); err != nil {
				return fmt.Errorf("field '%s': %w", field.Name, err)
			}
		}
	}
	return nil
}

// setFieldValue 设置结构体字段值
// fieldVal: 结构体字段值
// value: 要设置的值
func setFieldValue(fieldVal reflect.Value, value interface{}) error {
	if !fieldVal.CanSet() {
		return fmt.Errorf("cannot set field value")
	}

	valueVal := reflect.ValueOf(value)

	// 如果类型直接匹配，直接设置
	if valueVal.Type().AssignableTo(fieldVal.Type()) {
		fieldVal.Set(valueVal)
		return nil
	}

	// 处理基本类型转换
	if valueVal.Kind() == reflect.Float64 {
		floatVal := valueVal.Float()
		switch fieldVal.Kind() {
		case reflect.Int:
			fieldVal.SetInt(int64(floatVal))
		case reflect.Int32:
			fieldVal.SetInt(int64(floatVal))
		case reflect.Int64:
			fieldVal.SetInt(int64(floatVal))
		case reflect.Float32:
			fieldVal.SetFloat(floatVal)
		case reflect.Float64:
			fieldVal.SetFloat(floatVal)
		case reflect.Bool:
			fieldVal.SetBool(floatVal != 0)
		default:
			return fmt.Errorf("unsupported type conversion from float64 to %s", fieldVal.Kind())
		}
		return nil
	}

	// 处理切片类型转换
	if valueVal.Kind() == reflect.Slice && fieldVal.Kind() == reflect.Slice {
		return handleSliceConversion(fieldVal, value)
	}

	return fmt.Errorf("type mismatch: cannot assign %T to %s", value, fieldVal.Type())
}

// handleSliceConversion 处理切片类型转换
func handleSliceConversion(fieldVal reflect.Value, value interface{}) error {
	valueVal := reflect.ValueOf(value)
	
	// 确保都是切片类型
	if valueVal.Kind() != reflect.Slice || fieldVal.Kind() != reflect.Slice {
		return fmt.Errorf("both source and target must be slice types")
	}

	// 创建目标切片
	targetType := fieldVal.Type()
	targetSlice := reflect.MakeSlice(targetType, valueVal.Len(), valueVal.Len())

	// 逐个元素转换
	for i := 0; i < valueVal.Len(); i++ {
		elem := valueVal.Index(i)
		targetElem := targetSlice.Index(i)

		// 处理 interface{} 到具体类型的转换
		if elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}

		// 字符串切片转换
		if targetType.Elem().Kind() == reflect.String {
			if elem.Kind() == reflect.String {
				targetElem.SetString(elem.String())
			} else {
				// 尝试将其他类型转换为字符串
				str := fmt.Sprintf("%v", elem.Interface())
				targetElem.SetString(str)
			}
		} else if elem.Type().AssignableTo(targetType.Elem()) {
			// 直接赋值
			targetElem.Set(elem)
		} else {
			return fmt.Errorf("cannot convert element type %T to %s", elem.Interface(), targetType.Elem())
		}
	}

	fieldVal.Set(targetSlice)
	return nil
}
