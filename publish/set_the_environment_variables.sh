echo $jobs
for job in $jobs
do
    echo $job
	if [[ $job =~ "k8s" ]]; then
    	echo "hi"
    	echo x86_k8s_svc_jenkins_job="matan_test" > /root/git/ibm-block-csi-operator/publish/env.propert
    elif [[ $job =~ "production_rhel_svc" ]]; then
    	echo x86_ocp_svc_jenkins_job=$job >> /root/git/ibm-block-csi-operator/publish/env.propert
    else
    	echo z_ocp_svc_jenkins_job=$job >> $WORKSPACE/env.propert
    fi
done