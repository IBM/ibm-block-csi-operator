#!/usr/bin/python3
import requests
import json
import os
import time
import jenkins

jenkins_url = 'http://csijen.xiv.ibm.com:8080'
jenkins_user_name = os.getenv('USERNAME')
jenkins_password = os.getenv('PASSWORD')
latest_jobs_name = {}
jobs_status = {}
last_builds = {}


def get_all_jenkins_jobs_names():
    all_jenkins_jobs_rest_api_url = '{0}/api/json?tree=jobs[name]'.format(
        jenkins_url)
    message = requests.get(
        all_jenkins_jobs_rest_api_url, auth=(
            jenkins_user_name, jenkins_password))
    message = json.loads(message.text)
    return message['jobs']


def get_list_of_latest_jobs(
        latest_job,
        latest_major_version,
        latest_minor_version,
        job_name,
        latest_jobs_name):
    job_name_splited = job_name.split(".")
    temp_major_version = int(job_name_splited[0][-1])
    if temp_major_version >= latest_major_version[0]:
        latest_major_version[0] = temp_major_version
        if int(job_name_splited[1]) >= latest_minor_version[0]:
            latest_minor_version[0] = int(job_name_splited[1])
            latest_jobs_name[latest_job] = job_name


def get_the_desired_jobs(jobs):
    latest_k8s_major_version = [0]
    latest_k8s_minor_version = [0]
    latest_x86_ocp_major_version = [0]
    latest_x86_ocp_minor_version = [0]
    latest_z_ocp_major_version = [0]
    latest_z_ocp_minor_version = [0]
    latest_power_ocp_major_version = [0]
    latest_power_ocp_minor_version = [0]

    for job in jobs:

        if 'staging_svc' in job['name'].lower() and 'k8s' in job['name'].lower(
        ) and 'deprecated' not in job['name'].lower():
            get_list_of_latest_jobs(
                "latest_x86_k8s_job",
                latest_k8s_major_version,
                latest_k8s_minor_version,
                job['name'],
                latest_jobs_name)

        if 'production_rhel_svc' in job['name'].lower(
        ) and 'ocp' in job['name'].lower() and 'deprecated' not in job['name'].lower():
            get_list_of_latest_jobs(
                "latest_x86_ocp_job",
                latest_x86_ocp_major_version,
                latest_x86_ocp_minor_version,
                job['name'],
                latest_jobs_name)

        if 'production_z' in job['name'].lower() and 'svc' in job['name'].lower(
        ) and 'deprecated' not in job['name'].lower():
            get_list_of_latest_jobs(
                "latest_z_ocp_job",
                latest_z_ocp_major_version,
                latest_z_ocp_minor_version,
                job['name'],
                latest_jobs_name)

        if 'production_power_ocp' in job['name'].lower(
        ) and 'deprecated' not in job['name'].lower():
            get_list_of_latest_jobs(
                "latest_power_ocp_job",
                latest_power_ocp_major_version,
                latest_power_ocp_minor_version,
                job['name'],
                latest_jobs_name)

    return latest_jobs_name


def execute_latest_jobs(latest_jobs_name, server):

    for key in latest_jobs_name:

        if 'x86' in key:
            server.build_job(latest_jobs_name[key],
                             {'CSI_OPERATOR_IMAGE': '',
                              'CSI_CONTROLLER_IMAGE': '',
                              'CSI_NODE_IMAGE': '',
                              'INSTALLATION_METHOD': 'UNMANAGED'})
            server.build_job(latest_jobs_name[key],
                             {'CSI_OPERATOR_IMAGE': '',
                              'CSI_CONTROLLER_IMAGE': '',
                              'CSI_NODE_IMAGE': '',
                              'INSTALLATION_METHOD': 'OLM'})
        else:
            server.build_job(latest_jobs_name[key],
                             {'CSI_OPERATOR_IMAGE': '',
                              'CSI_CONTROLLER_IMAGE': '',
                              'CSI_NODE_IMAGE': '',
                              'INSTALLATION_METHOD': 'UNMANAGED'})


def wait_until_all_jobs_finish_running(latest_jobs_name, server):
    time.sleep(5)

    for key in latest_jobs_name:
        last_builds['last_build_of_{0}'.format(latest_jobs_name[key])] = server.get_job_info(
            latest_jobs_name[key])['builds'][0]['number']

    for key in latest_jobs_name:
        wait_for_jobs_to_finish(last_builds['last_build_of_{0}'.format(
            latest_jobs_name[key])], latest_jobs_name[key], server, jobs_status, True)

    print(jobs_status)

    if None in jobs_status.values() or 'FAILURE' in jobs_status.values():
        raise Exception("There are jobs that didn't succeded")


def wait_for_jobs_to_finish(
        last_job_number,
        job_name,
        server,
        jobs_status,
        has_two_jobs=False):
    last_completed_job = server.get_job_info(
        job_name)['lastCompletedBuild']['number']

    while last_job_number != last_completed_job:
        time.sleep(30)
        print(last_job_number)
        last_completed_job = server.get_job_info(
            job_name)['lastCompletedBuild']['number']
        print(
            "last completed build numer of job {0} is {1}".format(
                job_name,
                str(last_completed_job)))

    print("job {0} has finished".format(job_name))
    if has_two_jobs:
        jobs_status["{0}_OLM".format(job_name)] = server.get_build_info(
            job_name, last_job_number)['result']
        jobs_status["{0}_UNMANAGED".format(job_name)] = server.get_build_info(
            job_name, last_job_number - 1)['result']
    else:
        jobs_status["{0}_OLM".format(job_name)] = server.get_build_info(
            job_name, last_job_number)['result']


def connect_to_jenkins():
    server = jenkins.Jenkins(
        jenkins_url,
        username=jenkins_user_name,
        password=jenkins_password)
    return server


if __name__ == "__main__":
    jobs = get_all_jenkins_jobs_names()
    latest_jobs_name = get_the_desired_jobs(jobs)
    server = connect_to_jenkins()
    execute_latest_jobs(latest_jobs_name, server)
    wait_until_all_jobs_finish_running(latest_jobs_name, server)
