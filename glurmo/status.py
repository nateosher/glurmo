# pyright: strict
import subprocess
import os
from collections import Counter
from .setup import parse_settings

def status(target_dir: str, username: str):
    #--------------------------------------------------
    # Settings dictionary
    #--------------------------------------------------
    maybe_settings_dict = parse_settings(os.path.join(target_dir, ".glurmo", "settings.toml"))
    if maybe_settings_dict == None:
        return None
    
    #--------------------------------------------------
    # Getting jobs
    #--------------------------------------------------
    settings_dict = maybe_settings_dict
    cur_jobs = bytes.decode(subprocess.check_output("squeue"), encoding='utf-8')
    cur_jobs = cur_jobs.split('\n')
    col_names = cur_jobs[0].split()
    cur_jobs = [job.split() for job in cur_jobs[1:]]
    user_index = col_names.index("USER")
    jobname_index = col_names.index("NAME")
    state_index = col_names.index("STATE")
    cur_jobs = [job for job in cur_jobs if (len(job) > 0 and 
                            job[user_index] == username and 
                            job[jobname_index].split("___")[0] == settings_dict["simulation"]["id"])]
    
    state_counts = Counter([job[state_index] for job in cur_jobs])
    for state, count in state_counts.items():
        print(state + " : " + str(count))
    print("")


