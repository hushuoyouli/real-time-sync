package parser

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/hushuoyouli/real-time-sync/behaviortree/iface"
	"github.com/hushuoyouli/real-time-sync/behaviortree/parser/register"
	"github.com/hushuoyouli/real-time-sync/util"
)

/* var configJsonObjCache map[string]map[string]interface{}
var jsonCachedMutex sync.Mutex */

/* func init() {
	configJsonObjCache = make(map[string]map[string]interface{})
} */

func getConfigJsonObj(bytes []byte) (map[string]interface{}, error) {
	/* 	jsonCachedMutex.Lock()
	   	defer jsonCachedMutex.Unlock() */

	//	md5 := util.Md5(bytes)
	/* 	key := string(bytes)
	   	if obj, ok := configJsonObjCache[key]; ok {
	   		return obj, nil
	   	} else { */
	var configJsonObj map[string]interface{} = make(map[string]interface{})
	if err := json.Unmarshal(bytes, &configJsonObj); err != nil {
		return nil, err
	} else {
		/* configJsonObjCache[key] = configJsonObj */
		return configJsonObj, nil
	}
	/* } */
}

func Deserialize(jsonTextBytes []byte, taskAddData *util.TaskAddData) (iface.ITask, error) {
	var configJsonObj map[string]interface{}
	var err error

	if configJsonObj, err = getConfigJsonObj(jsonTextBytes); err != nil {
		return nil, err
	}

	id2Task := make(map[int]iface.ITask)
	rootJsonTaskConfig, ok := configJsonObj["RootTask"]
	if !ok {
		return nil, errors.New("json文件缺少RootTask的配置")
	}

	task2VariableConfigs := make(map[iface.ITask]map[string]interface{})
	rootTask, err := initializeTask(rootJsonTaskConfig.(map[string]interface{}), id2Task, task2VariableConfigs)
	if err != nil {
		return nil, err
	}

	detachedTasksConfigs, ok := configJsonObj["DetachedTasks"]
	if !ok {
		detachedTasksConfigs = make([]interface{}, 0)
	}

	detachedTasks := make([]iface.ITask, 0)
	for _, detachedTasksConfig := range detachedTasksConfigs.([]interface{}) {
		detachedTask, err := initializeTask(detachedTasksConfig.(map[string]interface{}), id2Task, task2VariableConfigs)
		if err != nil {
			return nil, err
		}

		detachedTasks = append(detachedTasks, detachedTask)
	}

	for task, variableMap := range task2VariableConfigs {
		if len(variableMap) > 0 {
			variableConfigs := make(map[string]interface{})

			for k, v := range variableMap {
				ks := strings.Split(k, ",")
				variableConfigs[ks[len(ks)-1]] = v
			}

			if err := task.SetVariables(variableConfigs); err != nil {
				return nil, err
			}

			//	以下部分有效率问题，采用虚函数进行优化
			/* 			if err := setVariableForTask(task, variableMap); err != nil {
				return nil, err
			} */

			//task.Print()
			//fmt.Println(task)
			//break
		}
	}

	initializeParentTask(rootTask, taskAddData)

	for _, disdetachedTask := range detachedTasks {
		initializeParentTask(disdetachedTask, taskAddData)
	}
	//	fmt.Println(detachedTasksConfig)

	return rootTask, nil
}

func initializeParentTask(task iface.ITask, taskAddData *util.TaskAddData) {
	task.SetParent(taskAddData.Parent)
	task.SetOwner(taskAddData.Owner)
	task.SetUnit(taskAddData.Unit)

	//if util.IsImplementsIParentTask(reflect.TypeOf(task).Elem()) {
	if task.IsImplementsIParentTask() {
		parent := taskAddData.Parent

		parentTask := task.(iface.IParentTask)
		taskAddData.Parent = parentTask

		for _, child := range parentTask.Children() {
			initializeParentTask(child, taskAddData)
		}

		taskAddData.Parent = parent
	}

}

func setVariableForTask(task iface.ITask, variableMap map[string]interface{}) error {
	ty := reflect.TypeOf(task).Elem()
	val := reflect.ValueOf(task).Elem()

	return setVariableForStruct(ty, val, variableMap)
}

/* var fieldMapCached = make(map[reflect.Type]map[string]reflect.StructField)
var valueMapCached = make(map[reflect.Type]map[string]reflect.Value)
var collectStructFieldMutex sync.Mutex */

func setVariableForStruct(ty reflect.Type, val reflect.Value, variableMap map[string]interface{}) error {
	/* 	var fieldMap map[string]reflect.StructField
	   	var valueMap map[string]reflect.Value */

	valueMap := make(map[string]reflect.Value)
	fieldMap := make(map[string]reflect.StructField)
	if err := collectStructField(ty, fieldMap, val, valueMap); err != nil {
		return err
	}

	/* if err := func(ty reflect.Type) error {
		collectStructFieldMutex.Lock()
		defer collectStructFieldMutex.Unlock()

		if _, ok := fieldMapCached[ty]; ok {
			fieldMap = fieldMapCached[ty]
			valueMap = valueMapCached[ty]
			return nil
		} else {
			valueMap = make(map[string]reflect.Value)
			fieldMap = make(map[string]reflect.StructField)
			if err := collectStructField(ty, fieldMap, val, valueMap); err != nil {
				return err
			} else {
				fieldMapCached[ty] = fieldMap
				valueMapCached[ty] = valueMap
				return nil
			}
		}
	}(ty); err != nil {
		return err
	} */

	for fieldName, value := range variableMap {
		fieldNames := strings.Split(fieldName, ",")
		if len(fieldNames) > 1 {
			fieldName = fieldNames[1]
		}

		fieldTy, ok := fieldMap[fieldName]
		if !ok {
			return errors.New("字段:" + fieldName + "不存在")
		}
		fieldVal := valueMap[fieldName]
		if err := setVariable(fieldTy, fieldVal, value); err != nil {
			return err
		}
	}

	return nil
}

func setVariable(fieldTy reflect.StructField, fieldVal reflect.Value, jsonVal interface{}) error {
	val, err := valFromJsonval(fieldTy.Type, jsonVal)
	if err != nil {
		return err
	}

	fieldVal.Set(reflect.ValueOf(val))
	return nil
}

func valFromJsonval(ty reflect.Type, jsonVal interface{}) (interface{}, error) {
	transform, ok := register.ValFromJsonObjTranforms[ty]
	if ok {
		return transform(jsonVal)
	}

	switch ty.Kind() {
	//	整数
	case reflect.Int32:
		return int32(jsonVal.(float64)), nil
	case reflect.Int64:
		return int64(jsonVal.(float64)), nil
	case reflect.Int:
		return int(jsonVal.(float64)), nil
	case reflect.Int8:
		return int8(jsonVal.(float64)), nil
	case reflect.Int16:
		return int16(jsonVal.(float64)), nil

	case reflect.Uint32:
		return uint32(jsonVal.(float64)), nil
	case reflect.Uint64:
		return uint64(jsonVal.(float64)), nil
	case reflect.Uint:
		return uint(jsonVal.(float64)), nil
	case reflect.Uint8:
		return uint8(jsonVal.(float64)), nil
	case reflect.Uint16:
		return uint16(jsonVal.(float64)), nil

	case reflect.Float32:
		return float32(jsonVal.(float64)), nil

	case reflect.Float64:
		return float64(jsonVal.(float64)), nil

	case reflect.String:
		return jsonVal, nil
	case reflect.Bool:
		return jsonVal, nil
	case reflect.Struct:
		value := reflect.New(ty)
		if err := setVariableForStruct(ty, value.Elem(), jsonVal.(map[string]interface{})); err != nil {
			return nil, err
		}

		return value.Elem().Interface(), nil
	case reflect.Array:
		arrayVals := reflect.ValueOf(jsonVal)
		if ty.Len() != arrayVals.Len() {
			return nil, errors.New("数组的长度与配置的不一样")
		}

		resultVal := reflect.New(ty).Elem()
		for i := 0; i < arrayVals.Len(); i++ {
			childVal, err := valFromJsonval(ty.Elem(), arrayVals.Index(i).Interface())

			if err != nil {
				return nil, err
			}

			resultVal.Index(i).Set(reflect.ValueOf(childVal))
		}

		return resultVal.Interface(), nil
	case reflect.Slice:
		sliceVals := reflect.ValueOf(jsonVal)
		sliceResult := reflect.MakeSlice(ty, sliceVals.Len(), sliceVals.Len())

		for i := 0; i < sliceVals.Len(); i++ {
			//for i, sliceVal := range sliceVals {
			childVal, err := valFromJsonval(ty.Elem(), sliceVals.Index(i).Interface())

			if err != nil {
				return nil, err
			}

			sliceResult.Index(i).Set(reflect.ValueOf(childVal))
		}
		return sliceResult.Interface(), nil
	case reflect.Pointer:
		elemPrt := reflect.New(ty.Elem())
		elemVal, err := valFromJsonval(ty.Elem(), jsonVal)
		if err != nil {
			return nil, err
		}

		elemPrt.Elem().Set(reflect.ValueOf(elemVal))
		return elemPrt.Interface(), nil
	default:
		return jsonVal, nil
	}
}

func collectStructField(ty reflect.Type, fieldMap map[string]reflect.StructField, val reflect.Value, valueMap map[string]reflect.Value) error {
	for i := 0; i < ty.NumField(); i++ {
		if val.Field(i).CanSet() {
			if ty.Field(i).Anonymous {
				if err := collectStructField(ty.Field(i).Type, fieldMap, val.Field(i), valueMap); err != nil {
					return err
				}
			} else {
				tagName := strings.TrimSpace(ty.Field(i).Tag.Get("behaviortree"))
				if tagName != "" {
					if tagName != "-" {
						fieldMap[tagName] = ty.Field(i)
						valueMap[tagName] = val.Field(i)
					}
				} else {
					tagName = ty.Field(i).Name
					fieldMap[tagName] = ty.Field(i)
					valueMap[tagName] = val.Field(i)
				}
			}
		}
	}

	return nil
}

func initializeTask(config map[string]interface{}, id2Task map[int]iface.ITask, task2VariableConfigs map[iface.ITask]map[string]interface{}) (iface.ITask, error) {
	correspondingType := config["Type"].(string)

	task, err := register.NewTask(correspondingType)
	if err != nil {
		return nil, err
	}

	variableConfigs := make(map[string]interface{}, 0)

	for k, v := range config {
		switch k {
		case "Type":
			task.SetCorrespondingType(v.(string))
		case "Children":
		case "Name":
			task.SetName(v.(string))
		case "ID":
			task.SetID(int(v.(float64)))
		case "Instant":
			task.SetIsInstant(v.(bool))
		case "Disabled":
			task.SetDisabled(v.(bool))
		case "BehaviorDesigner.Runtime.Tasks.AbortType,abortType":
			compositeTask, ok := task.(iface.IComposite)
			if ok {
				switch v.(string) {
				case "Both":
					compositeTask.SetAbortType(iface.Both)
				case "Self":
					compositeTask.SetAbortType(iface.Self)
				case "LowerPriority":
					compositeTask.SetAbortType(iface.LowerPriority)
				default:
					compositeTask.SetAbortType(iface.None)
				}
			}

		default:
			variableConfigs[k] = v
		}
	}

	taskID := task.ID()
	if taskID == 0 {
		return nil, errors.New("任务缺少配置参数ID")
	}

	_, ok := id2Task[taskID]
	if ok {
		return nil, errors.New("任务ID:" + strconv.FormatInt(int64(taskID), 10) + "有冲突")
	}

	id2Task[taskID] = task
	task2VariableConfigs[task] = variableConfigs

	parentTask, ok := task.(iface.IParentTask)

	if ok {
		childrenConfigs, ok := config["Children"]
		if ok {
			for _, childConfig := range childrenConfigs.([]interface{}) {
				childTask, err := initializeTask(childConfig.(map[string]interface{}), id2Task, task2VariableConfigs)
				if err != nil {
					return nil, err
				}
				childTask.SetParent(parentTask)

				//if !childTask.Disabled() {
				parentTask.AddChild(childTask)
				//}
			}
		}
	}

	return task, nil
}
