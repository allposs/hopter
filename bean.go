package hopter

import (
	"reflect"
)

// BeanFactory Bean工厂
type BeanFactory struct {
	beans []any
}

// NewBeanFactory 创建Bean工厂
func NewBeanFactory() *BeanFactory {
	bf := &BeanFactory{beans: make([]any, 0)}
	bf.beans = append(bf.beans, bf)
	return bf
}

// set 往内存中塞入bean
func (b *BeanFactory) set(beans ...any) {
	for _, p := range beans {
		obj := reflect.TypeOf(p)
		if obj.Kind() == reflect.Func {
			conf := b.get(new(Endpoint)).(*Endpoint)
			b.beans = append(b.beans, p.(func(conf Endpoint) any)(*conf))
			continue
		}
		b.beans = append(b.beans, p)
	}
}

// get 外部使用
func (b *BeanFactory) get(bean any) any {
	return b.find(reflect.TypeOf(bean))
}

// find 得到内存中预先设置好的bean对象
func (b *BeanFactory) find(t reflect.Type) any {
	for _, p := range b.beans {
		if t == reflect.TypeOf(p) {
			return p
		}
	}
	return nil
}

// Inject 把bean注入到控制器中
func (b *BeanFactory) Inject(object any) {
	vObject := reflect.ValueOf(object)
	if vObject.Kind() == reflect.Ptr {
		//由于不是控制器 ，所以传过来的值 不一定是指针。因此要做判断
		vObject = vObject.Elem()
	}
	for i := 0; i < vObject.NumField(); i++ {
		f := vObject.Field(i)
		if f.Kind() != reflect.Ptr || !f.IsNil() {
			continue
		}
		if p := b.find(f.Type()); p != nil && f.CanInterface() {
			f.Set(reflect.New(f.Type().Elem()))
			f.Elem().Set(reflect.ValueOf(p).Elem())
		}
	}
}

// Beans Bean注册
func (w *Web) Beans(beans ...any) *Web {
	w.beanFactory.set(beans...)
	return w
}
