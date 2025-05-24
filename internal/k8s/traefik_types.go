package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// IngressRoute defines the Traefik IngressRoute CRD
type IngressRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IngressRouteSpec `json:"spec"`
}

// IngressRouteSpec defines the desired state of IngressRoute
type IngressRouteSpec struct {
	EntryPoints []string          `json:"entryPoints,omitempty"`
	Routes      []Route           `json:"routes"`
	TLS         *TLS              `json:"tls,omitempty"`
}

// Route defines a route in an IngressRoute
type Route struct {
	Match    string     `json:"match"`
	Kind     string     `json:"kind"`
	Services []Service  `json:"services,omitempty"`
}

// Service defines a service in a route
type Service struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

// TLS defines the TLS section of an IngressRoute
type TLS struct {
	CertResolver string `json:"certResolver,omitempty"`
}

// IngressRouteList defines a list of IngressRoute
type IngressRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IngressRoute `json:"items"`
}

// DeepCopyInto copies all properties from this object into another object
func (in *IngressRoute) DeepCopyInto(out *IngressRoute) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
}

// DeepCopy copies the receiver, creating a new IngressRoute
func (in *IngressRoute) DeepCopy() *IngressRoute {
	if in == nil {
		return nil
	}
	out := new(IngressRoute)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject returns a generically copied version of the object
func (in *IngressRoute) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto copies all properties from this object into another object
func (in *IngressRouteList) DeepCopyInto(out *IngressRouteList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]IngressRoute, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy copies the receiver, creating a new IngressRouteList
func (in *IngressRouteList) DeepCopy() *IngressRouteList {
	if in == nil {
		return nil
	}
	out := new(IngressRouteList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject returns a generically copied version of the object
func (in *IngressRouteList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// GroupVersionResource returns the GroupVersionResource for IngressRoute
func IngressRouteGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "traefik.io",
		Version:  "v1alpha1",
		Resource: "ingressroutes",
	}
}