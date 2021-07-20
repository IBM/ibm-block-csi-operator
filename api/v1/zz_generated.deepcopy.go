// +build !ignore_autogenerated

/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CSISidecar) DeepCopyInto(out *CSISidecar) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CSISidecar.
func (in *CSISidecar) DeepCopy() *CSISidecar {
	if in == nil {
		return nil
	}
	out := new(CSISidecar)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IBMBlockCSI) DeepCopyInto(out *IBMBlockCSI) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IBMBlockCSI.
func (in *IBMBlockCSI) DeepCopy() *IBMBlockCSI {
	if in == nil {
		return nil
	}
	out := new(IBMBlockCSI)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IBMBlockCSI) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IBMBlockCSIControllerSpec) DeepCopyInto(out *IBMBlockCSIControllerSpec) {
	*out = *in
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(corev1.Affinity)
		(*in).DeepCopyInto(*out)
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]corev1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IBMBlockCSIControllerSpec.
func (in *IBMBlockCSIControllerSpec) DeepCopy() *IBMBlockCSIControllerSpec {
	if in == nil {
		return nil
	}
	out := new(IBMBlockCSIControllerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IBMBlockCSIList) DeepCopyInto(out *IBMBlockCSIList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]IBMBlockCSI, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IBMBlockCSIList.
func (in *IBMBlockCSIList) DeepCopy() *IBMBlockCSIList {
	if in == nil {
		return nil
	}
	out := new(IBMBlockCSIList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IBMBlockCSIList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IBMBlockCSINodeSpec) DeepCopyInto(out *IBMBlockCSINodeSpec) {
	*out = *in
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(corev1.Affinity)
		(*in).DeepCopyInto(*out)
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]corev1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IBMBlockCSINodeSpec.
func (in *IBMBlockCSINodeSpec) DeepCopy() *IBMBlockCSINodeSpec {
	if in == nil {
		return nil
	}
	out := new(IBMBlockCSINodeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IBMBlockCSISpec) DeepCopyInto(out *IBMBlockCSISpec) {
	*out = *in
	in.Controller.DeepCopyInto(&out.Controller)
	in.Node.DeepCopyInto(&out.Node)
	if in.Sidecars != nil {
		in, out := &in.Sidecars, &out.Sidecars
		*out = make([]CSISidecar, len(*in))
		copy(*out, *in)
	}
	if in.ImagePullSecrets != nil {
		in, out := &in.ImagePullSecrets, &out.ImagePullSecrets
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IBMBlockCSISpec.
func (in *IBMBlockCSISpec) DeepCopy() *IBMBlockCSISpec {
	if in == nil {
		return nil
	}
	out := new(IBMBlockCSISpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IBMBlockCSIStatus) DeepCopyInto(out *IBMBlockCSIStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IBMBlockCSIStatus.
func (in *IBMBlockCSIStatus) DeepCopy() *IBMBlockCSIStatus {
	if in == nil {
		return nil
	}
	out := new(IBMBlockCSIStatus)
	in.DeepCopyInto(out)
	return out
}
