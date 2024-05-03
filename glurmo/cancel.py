# pyright: strict
from typing import List
import subprocess
import os
from .setup import parse_settings

# TODO: add state of simulations to cancel
def cancel(next_action: List[str], target_dir: str, username: str):
    maybe_settings_dict = parse_settings(os.path.join(target_dir, ".glurmo", "settings.toml"))
    if maybe_settings_dict == None:
        return None
    settings_dict = maybe_settings_dict
    job_id = settings_dict["simulation"]["id"]
    #--------------------------------------------------
    # Input checking
    #--------------------------------------------------
    if len(next_action) != 2:
        print("Usage: c [# sims to cancel]")
        return None

    try:
        n_to_cancel = int(next_action[1])
    except ValueError:
        print("Error: invalid number of simulations to cancel: " + next_action[1])
        return None

    if n_to_cancel <= 0:
        print("Error: number of simulations to cancel should be greater than 0")
        return None

    #--------------------------------------------------
    # Getting job ids to cancel
    #--------------------------------------------------
    cur_jobs = bytes.decode(subprocess.check_output("squeue"), encoding='utf-8')
    cur_jobs = cur_jobs.split('\n')
    col_names = cur_jobs[0].split()
    cur_jobs = [job.split() for job in cur_jobs[1:]]
    user_index = col_names.index("USER")
    jobname_index = col_names.index("NAME")
    jobid_index = col_names.index("JOBID")
    cur_jobs = [job for job in cur_jobs if (len(job) > 0 and 
                            job[user_index] == username and 
                            job[jobname_index].split("___")[0] == job_id)]
    cur_job_ids = [job[jobid_index] for job in cur_jobs]

    #--------------------------------------------------
    # Cancel jobs
    #--------------------------------------------------
    n_canceled = 0
    i = -1
    while n_canceled < n_to_cancel and i < len(cur_job_ids) - 1:
        i += 1
        _ = bytes.decode(
            subprocess.run(["scancel", cur_job_ids[i]], stdout=subprocess.PIPE).stdout,
            encoding = "utf-8"
        )
        n_canceled += 1

    
    print("Successfully canceled " + str(n_canceled) + " jobs")