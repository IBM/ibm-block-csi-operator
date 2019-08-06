package fakeserver_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/IBM/ibm-block-csi-driver-operator/pkg/node/fakeserver"
	pb "github.com/IBM/ibm-block-csi-driver-operator/pkg/node/nodeagent"
)

var _ = Describe("Store", func() {

	AfterEach(func() {
		ClearAll()
	})

	Context("test store one response", func() {
		var req *pb.GetNodeInfoRequest
		var res *pb.GetNodeInfoReply

		BeforeEach(func() {
			req = &pb.GetNodeInfoRequest{Name: "node-1"}
			res = &pb.GetNodeInfoReply{Node: &pb.Node{Name: "node-1", Iqns: []string{"iqn.xxx"}}}
		})

		It("should get right res after store in same gorutine", func() {
			StoreResponse("GetNodeInfo", req, res)
			resFromStore := LoadResponse("GetNodeInfo", req).(*pb.GetNodeInfoReply)
			Expect(resFromStore).NotTo(BeNil())
			Expect(resFromStore.GetNode()).To(Equal(res.GetNode()))

			// get again
			res1FromStore := LoadResponse("GetNodeInfo", req).(*pb.GetNodeInfoReply)
			Expect(res1FromStore).NotTo(BeNil())
			Expect(res1FromStore.GetNode()).To(Equal(res.GetNode()))

		})
		It("should get right res after store in another gorutine", func() {
			done := make(chan bool)
			go func() {
				StoreResponse("GetNodeInfo", req, res)
				done <- true
			}()
			<-done

			resFromStore := LoadResponse("GetNodeInfo", req).(*pb.GetNodeInfoReply)
			Expect(resFromStore).NotTo(BeNil())
			Expect(resFromStore.GetNode()).To(Equal(res.GetNode()))

		})
	})

	Context("test store more than one response", func() {
		var req *pb.GetNodeInfoRequest
		var res1, res2 *pb.GetNodeInfoReply

		BeforeEach(func() {
			req = &pb.GetNodeInfoRequest{Name: "node-1"}
			res1 = &pb.GetNodeInfoReply{Node: &pb.Node{Name: "node-1", Iqns: []string{"iqn.xxx"}}}
			res2 = &pb.GetNodeInfoReply{Node: &pb.Node{Name: "node-1", Iqns: []string{"iqn.yyy"}}}
			StoreResponse("GetNodeInfo", req, res1)
		})

		It("should get right res after store in same gorutine", func() {
			StoreResponse("GetNodeInfo", req, res2)
			res1FromStore := LoadResponse("GetNodeInfo", req).(*pb.GetNodeInfoReply)
			Expect(res1FromStore).NotTo(BeNil())
			Expect(res1FromStore.GetNode().GetIqns()).To(Equal(res1.GetNode().GetIqns()))

			res2FromStore := LoadResponse("GetNodeInfo", req).(*pb.GetNodeInfoReply)
			Expect(res2FromStore).NotTo(BeNil())
			Expect(res2FromStore.GetNode().GetIqns()).To(Equal(res2.GetNode().GetIqns()))

		})

		It("should get right res after store in different order", func() {
			// get res1 first
			res1FromStore := LoadResponse("GetNodeInfo", req).(*pb.GetNodeInfoReply)
			Expect(res1FromStore).NotTo(BeNil())
			Expect(res1FromStore.GetNode().GetIqns()).To(Equal(res1.GetNode().GetIqns()))

			// store res2
			StoreResponse("GetNodeInfo", req, res2)

			// get res2 after store
			res2FromStore := LoadResponse("GetNodeInfo", req).(*pb.GetNodeInfoReply)
			Expect(res2FromStore).NotTo(BeNil())
			Expect(res2FromStore.GetNode().GetIqns()).To(Equal(res2.GetNode().GetIqns()))

			// get res1 again
			res3FromStore := LoadResponse("GetNodeInfo", req).(*pb.GetNodeInfoReply)
			Expect(res3FromStore).NotTo(BeNil())
			Expect(res3FromStore.GetNode().GetIqns()).To(Equal(res1.GetNode().GetIqns()))
		})

		It("should get right res after store in another gorutine", func() {
			res1FromStore := LoadResponse("GetNodeInfo", req).(*pb.GetNodeInfoReply)
			Expect(res1FromStore).NotTo(BeNil())
			Expect(res1FromStore.GetNode().GetIqns()).To(Equal(res1.GetNode().GetIqns()))

			done := make(chan bool)
			go func() {
				StoreResponse("GetNodeInfo", req, res2)
				done <- true
			}()
			<-done

			res2FromStore := LoadResponse("GetNodeInfo", req).(*pb.GetNodeInfoReply)
			Expect(res2FromStore).NotTo(BeNil())
			Expect(res2FromStore.GetNode().GetIqns()).To(Equal(res2.GetNode().GetIqns()))

		})
	})
})
