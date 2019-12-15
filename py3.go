package py3

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/DataDog/go-python3"
	log "github.com/sirupsen/logrus"
)

var (
	module *python3.PyObject
	locker sync.Mutex
)

type resp [][]float64

func Init(m string) *python3.PyObject {
	python3.Py_Initialize()
	// defer python3.Py_Finalize()
	if !python3.Py_IsInitialized() {
		panic("Error initializing the python interpreter")
	}
	module = python3.PyImport_ImportModule(m)
	if module == nil {
		panic("module is nil")
	}
	return module
}

func pythonRepr(o *python3.PyObject) (string, error) {
	if o == nil {
		return "", fmt.Errorf("object is nil")
	}

	s := o.Repr()
	if s == nil {
		python3.PyErr_Clear()
		return "", fmt.Errorf("failed to call Repr object method")
	}
	defer s.DecRef()

	return python3.PyUnicode_AsUTF8(s), nil
}

func GoPyFunc(funcname string, args ...float64) []float64 {
	fname := module.GetAttrString(funcname)
	if fname == nil {
		log.Fatalf("could not getattr(%s, '%s')\n", funcname, funcname)
	}
	log.Debugf("GoPyFunc, %s", funcname)

	pyargs := ToPyTuple(args...)
	pyparams := ToPyDict(args...)
	log.Infof("fname:%+v", *fname)
	log.Infof("pyargs:%+v", pyargs)
	log.Infof("pyparams:%+v", pyparams)

	out := fname.Call(pyargs, pyparams)
	log.Infof("out:%+v", out)

	s, err := pythonRepr(out)
	log.Infof("str:%s, err:%+v", s, err)

	// log.Infof("out string:%s", python3.PyBytes_AsString(out.Bytes()))

	return nil
}

func ToPyTuple(vs ...float64) *python3.PyObject {
	args := python3.PyTuple_New(len(vs))
	for i, v := range vs {
		python3.PyTuple_SetItem(args, i, python3.PyFloat_FromDouble(v))
	}
	return args
}

func ToPyDict(vs ...float64) *python3.PyObject {
	args := python3.PyDict_New()
	for i, v := range vs {
		python3.PyDict_SetItem(
			args,
			python3.PyUnicode_FromString(strconv.FormatInt(int64(i), 10)), // ?
			python3.PyFloat_FromDouble(v),
		)
	}
	return args
}

func GoPyFuncV2(funcname string, args [][]float64, params map[string]int32) ([][]float64, error) {
	locker.Lock()
	defer locker.Unlock()
	log.Debugf("GoPyFuncV2, %s", funcname)

	var out *python3.PyObject
	fname := module.GetAttrString(funcname)
	if fname == nil {
		err := fmt.Errorf("could not getattr(%s, '%s')", funcname, funcname)
		log.Error(err)
		return nil, err
	}
	log.Debugf("GoPyFuncV2, %s", funcname)

	pyargs := ToPyListV2(args)
	pyparams := ToPyDictV2(params)
	log.Infof("fname:%+v", *fname)
	log.Infof("pyargs:%+v", pyargs)
	log.Infof("pyparams:%+v", pyparams)

	out = fname.Call(pyargs, pyparams)
	log.Infof("out:%+v", out)
	s, err := pythonRepr(out)
	log.Infof("str:%s, err:%+v", s, err)
	s = strings.Trim(s, "'")
	fmt.Println(s[:len(s)])
	var r resp
	err = json.Unmarshal([]byte(s), &r)
	// err = json.Unmarshal(goutils.ToByte(s), &r)
	if err != nil {
		log.Errorf("err:%+v", err)
		return nil, err
	}
	log.Infof("resp:%+v", r)
	return [][]float64(r), nil
}

func ToPyListV2(input [][]float64) *python3.PyObject {
	args := python3.PyTuple_New(len(input))
	for i, it := range input {
		subargs := python3.PyList_New(0)
		for _, jt := range it {
			python3.PyList_Append(subargs, python3.PyFloat_FromDouble(jt))
		}
		python3.PyTuple_SetItem(args, i, subargs)
	}
	return args
}

func ToPyDictV2(vs map[string]int32) *python3.PyObject {
	args := python3.PyDict_New()
	for k, v := range vs {
		python3.PyDict_SetItem(
			args,
			python3.PyUnicode_FromString(k),
			python3.PyLong_FromLong(int(v)),
		)
	}
	return args
}
