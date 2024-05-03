# pyright: strict
import os
import re
from typing import Dict, Union
def setup(target_dir: str):
    #--------------------------------------------------
    # Check for proper configuration
    #--------------------------------------------------

    # check for .glurmo
    dir_contents = set(os.listdir(target_dir))
    if len(dir_contents) > 1:
        print("Error: target directory should only contain `.glurmo` subdirectory")
        return None
    if ".glurmo" not in dir_contents:
        print("Error: target directory must contain `.glurmo` subdirectory")
        return None
    if not os.path.isdir(os.path.join(target_dir, ".glurmo")):
        print("Error: `.glurmo` should be a directory, not a file")
        return None
    
    # check .glurmo contents
    glurmo_path = os.path.join(target_dir, ".glurmo")
    glurmo_contents = set(os.listdir(glurmo_path))
    if "script_template" not in glurmo_contents:
        print("Error: `.glurmo` directory must contain `script_template` file")
        return None
    if not os.path.isfile(os.path.join(glurmo_path, "script_template")):
        print("Error: `.glurmo/script_template` must be a file")
    if "slurm_template" not in glurmo_contents:
        print("Error: `.glurmo` directory must contain `slurm_template` file")
        return None
    if not os.path.isfile(os.path.join(glurmo_path, "slurm_template")):
        print("Error: `.glurmo/slurm_template` must be a file")
    if "settings.toml" not in glurmo_contents:
        print("Error: `.glurmo` directory must contain `settings.toml` file")
        return None
    if not os.path.isfile(os.path.join(glurmo_path, "settings.toml")):
        print("Error: `.glurmo/settings.toml` must be a file")

    #--------------------------------------------------
    # Parse settings
    #--------------------------------------------------
    maybe_settings_dict = parse_settings(os.path.join(glurmo_path, "settings.toml"))
    if maybe_settings_dict == None:
        return None
    settings_dict = maybe_settings_dict
    
    #--------------------------------------------------
    # Create directory structure
    #--------------------------------------------------
    target_subdirs = ["scripts", "slurm", "results", "slurm_output"] 
    for dir in target_subdirs:
        os.mkdir(os.path.join(target_dir, dir))

    #--------------------------------------------------
    # Create scripts
    #--------------------------------------------------
    n_scripts = int(settings_dict["simulation"]["n_sim"])
    script_file = open(os.path.join(glurmo_path, "script_template"))
    script_template = script_file.read()
    script_file.close()
    for k, v in settings_dict["script"].items():
        script_template = re.sub(r"\{\{" + k + r"\}\}", v, script_template)
    
    for i in range(n_scripts):
        cur_script_path = os.path.join(target_dir, "scripts",
                    "script___" + str(i) + "." + settings_dict["simulation"]["script_extension"])
        cur_result_path = os.path.join(target_dir, 
            "results", "result___" + str(i) + "." + settings_dict["simulation"]["result_extension"]) 
        temp_settings = {
            "index" : str(i),
            "results_path" : cur_result_path
        }
        cur_script = script_template
        for k, v in temp_settings.items():
            cur_script = re.sub(r"\{\{" + k + r"\}\}", v, cur_script) 
        with open(cur_script_path, 'w') as f:
            f.write(cur_script)
    
    #--------------------------------------------------
    # Create slurm files
    #--------------------------------------------------
    slurm_file = open(os.path.join(glurmo_path, "slurm_template"))
    slurm_template = slurm_file.read()
    slurm_file.close()
    for k, v in settings_dict["slurm"].items():
        slurm_template = re.sub(r"\{\{" + k + r"\}\}", v, slurm_template)

    for i in range(n_scripts):
        cur_slurm_path = os.path.join(target_dir, "slurm", "slurm___" + str(i))
        cur_script_path = os.path.join(target_dir, "scripts",
                    "script___" + str(i) + "." + settings_dict["simulation"]["script_extension"])
        cur_output_path = os.path.join(target_dir, "slurm_output")
        temp_settings = {
            "job_id" : settings_dict["simulation"]["id"] + "___" + str(i),
            "script_path" : cur_script_path,
            "output_path" : cur_output_path,
            "index" : str(i)
        }
        cur_slurm = slurm_template
        for k, v in temp_settings.items():
            cur_slurm = re.sub(r"\{\{" + k + r"\}\}", v, cur_slurm) 
        with open(cur_slurm_path, 'w') as f:
            f.write(cur_slurm)

def parse_settings(settings_file_path: str) -> Union[Dict[str, Dict[str, str]], None]:
    #--------------------------------------------------
    # Parse settings
    #--------------------------------------------------
    settings_dict: Dict[str, Dict[str, str]] = {
        "simulation" : {},
        "script" : {},
        "slurm" : {}
    }

    with open(settings_file_path, 'r') as f:
        cur_settings_section = ''
        for line in f:
            line = line.split("#")[0]
            line = line.strip()
            if line == '': # ignore blank lines
                continue
            if line[0] == '#': # ignore comments
                continue
            if line[0] == '[':
                cur_settings_section = line[1:-1]
                continue

            split_settings = [s.strip() for s in line.split("=")]
            if len(split_settings) != 2:
                print("Error: malformed setting in `settings.toml`: " + line)
                return None
            if cur_settings_section not in settings_dict.keys():
                print("Error: all settings must be in [simulation], [script], or [slurm] section")
                print("The following setting is not: " + line)
                return None

            setting_name = split_settings[0]
            setting_value = split_settings[1]
            if setting_value[0] == '"' or setting_value[0] == "'":
                setting_value = setting_value[1:-1]

            settings_dict[cur_settings_section][setting_name] = setting_value

        return settings_dict