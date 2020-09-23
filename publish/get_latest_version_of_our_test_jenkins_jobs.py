#!/usr/bin/python3
import requests
import json
import os

jenkins_url = 'http://csijen.xiv.ibm.com:8080'
jenkins_user_name = os.getenv('USERNAME')
jenkins_password = os.getenv('PASSWORD')

def get_all_jenkins_jobs_names():
    all_jenkins_jobs_rest_api_url = '{0}/api/json?tree=jobs[name]'.format(jenkins_url)
    message = requests.get(all_jenkins_jobs_rest_api_url, auth=(jenkins_user_name, jenkins_password))
    message = json.loads(message.text)
    return message['jobs']

def get_list_of_latest_jobs(latest_job, latest_major_version, latest_minor_version, job_name):
        job_name_splited = job_name.split(".")
        temp_major_version = int(job_name_splited[0][-1])
        if temp_major_version >= latest_major_version[0]:
            latest_major_version[0] = temp_major_version
            if int(job_name_splited[1]) >= latest_minor_version[0]:
                latest_minor_version[0] = int(job_name_splited[1])
                latest_job[0] = job_name

def get_the_desired_jobs(jobs):
    latest_k8s_job = [""]
    latest_x86_ocp_job = [""]
    latest_z_ocp_job = [""]
    latest_k8s_major_version = [0]
    latest_k8s_minor_version = [0]
    latest_x86_ocp_major_version = [0]
    latest_x86_ocp_minor_version = [0]
    latest_z_ocp_major_version = [0]
    latest_z_ocp_minor_version = [0]

    for job in jobs:

        if 'staging_svc' in job['name'].lower() and 'k8s' in job['name'].lower() and 'deprecated' not in job['name'].lower():
            get_list_of_latest_jobs(latest_k8s_job, latest_k8s_major_version, latest_k8s_minor_version, job['name'])

        if 'production_rhel_svc' in job['name'].lower() and 'ocp' in job['name'].lower() and 'deprecated' not in job['name'].lower():
            get_list_of_latest_jobs(latest_x86_ocp_job, latest_x86_ocp_major_version, latest_x86_ocp_minor_version, job['name'])

        if 'production_z' in job['name'].lower() and  'svc' in job['name'].lower() and 'deprecated' not in job['name'].lower():
            get_list_of_latest_jobs(latest_z_ocp_job, latest_z_ocp_major_version, latest_z_ocp_minor_version, job['name'])
    
    print(latest_k8s_job[0])
    print(latest_x86_ocp_job[0])
    print(latest_z_ocp_job[0])

if __name__ == "__main__":
    jobs = get_all_jenkins_jobs_names()
    get_the_desired_jobs(jobs)