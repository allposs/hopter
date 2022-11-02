package web

import (
	"reflect"
)

// BeanFactory Bean工厂
type BeanFactory struct {
	beans []interface{}
}

// NewBeanFactory 创建Bean工厂
func NewBeanFactory() *BeanFactory {
	bf := &BeanFactory{beans: make([]interface{}, 0)}
	bf.beans = append(bf.beans, bf)
	return bf
}

// setBean 往内存中塞入bean
func (b *BeanFactory) setBean(beans ...interface{}) {
	for _, p := range beans {
		obj := reflect.TypeOf(p)
		if obj.Kind() == reflect.Func {
			conf := b.GetBean(new(Config)).(*Config)
			b.beans = append(b.beans, p.(func(conf Config) interface{})(*conf))
			continue
		}
		b.beans = append(b.beans, p)
	}
}

// GetBean 外部使用
func (b *BeanFactory) GetBean(bean interface{}) interface{} {
	return b.getBean(reflect.TypeOf(bean))
}

// getBean 得到 内存中预先设置好的bean对象
func (b *BeanFactory) getBean(t reflect.Type) interface{} {
	for _, p := range b.beans {
		if t == reflect.TypeOf(p) {
			return p
		}
	}
	return nil
}

// Inject 给外部用的 （后面还要改,这个方法不处理注解)
func (b *BeanFactory) Inject(object interface{}) {
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
		if p := b.getBean(f.Type()); p != nil && f.CanInterface() {
			f.Set(reflect.New(f.Type().Elem()))
			f.Elem().Set(reflect.ValueOf(p).Elem())
		}

	}
}

// inject 把bean注入到控制器中 (内部方法,用户控制器注入。并同时处理注解)
func (b *BeanFactory) inject(class Interface) {
	vClass := reflect.ValueOf(class).Elem()
	for i := 0; i < vClass.NumField(); i++ {
		f := vClass.Field(i)
		if f.Kind() != reflect.Ptr || !f.IsNil() {
			continue
		}
		if p := b.getBean(f.Type()); p != nil {
			f.Set(reflect.New(f.Type().Elem()))
			f.Elem().Set(reflect.ValueOf(p).Elem())
		}
	}
}
