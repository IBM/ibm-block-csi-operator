for job in $jobs
do
	if [ $job =~ "k8s"]; then
    	echo "hi"
    	echo x86_k8s_svc_jenkins_job=matan_test > $WORKSPACE/env.propert
    elif [ $job =~ "production_rhel_svc"]; then
    	echo x86_ocp_svc_jenkins_job=$job >> $WORKSPACE/env.propert
    else
    	echo z_ocp_svc_jenkins_job=$job >> $WORKSPACE/env.propert
    fi
done