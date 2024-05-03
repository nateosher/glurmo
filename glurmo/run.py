# pyright: strict
from typing import List
import subprocess
import os
from .setup import parse_settings

def run(next_action: List[str], target_dir: str, username: str):
    maybe_settings_dict = parse_settings(os.path.join(target_dir, ".glurmo", "settings.toml"))
    if maybe_settings_dict == None:
        return None
    settings_dict = maybe_settings_dict 
    #--------------------------------------------------
    # Input checking
    #--------------------------------------------------
    if len(next_action) != 2:
        print("Usage: r [# sims to run]")
        return None

    try:
        n_to_run = int(next_action[1])
    except ValueError:
        print("Error: invalid number of simulations to run: " + next_action[1])
        return None

    if n_to_run <= 0:
        print("Error: number of simulations to run should be greater than 0")
        return None
    
    #--------------------------------------------------
    # Figure out what not to submit
    #--------------------------------------------------
    # What jobs are currently running/pending
    # TODO: print with infinite job name characters
    cur_jobs = bytes.decode(subprocess.check_output("squeue"), encoding='utf-8')
    cur_jobs = cur_jobs.split('\n')
    col_names = cur_jobs[0].split()
    cur_jobs = [job.split() for job in cur_jobs[1:]]
    user_index = col_names.index("USER")
    jobname_index = col_names.index("NAME")
    cur_jobs = [job for job in cur_jobs if (len(job) > 0 and 
                            job[user_index] == username and 
                            job[jobname_index].split("___")[0] == settings_dict["simulation"]["id"])]
    cur_job_nums = [int(job[jobname_index].split("___")[1]) for job in cur_jobs]
    
    # What jobs are finished
    finished_jobs = os.listdir(os.path.join(target_dir, "results"))
    finished_job_nums = [int(job.split(".")[0].split("___")[1]) for job in finished_jobs]

    # Combine them into set
    already_submitted = set(cur_job_nums + finished_job_nums)

    #--------------------------------------------------
    # Submit jobs
    #--------------------------------------------------
    n_submitted = 0
    i = -1
    while n_submitted < n_to_run and i < int(settings_dict["simulation"]["n_sim"]) - 1:
        i += 1
        if i in already_submitted:
            continue
        cur_slurm_file = os.path.join(target_dir, "slurm", "slurm___" + str(i))
        command_output = bytes.decode(
            subprocess.run(["sbatch", cur_slurm_file], stdout=subprocess.PIPE).stdout,
            encoding = "utf-8"
        )
        if not command_output.startswith("Submitted batch job"):
            print("Error: could not submit job " + str(i) + ": " + command_output)
        else:
            n_submitted += 1

    
    print("Successfully submitted " + str(n_submitted) + " jobs")
