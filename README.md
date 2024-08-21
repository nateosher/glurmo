# glurmo

`glurmo` is a command-line simulation manager built for the `slurm` workload manager. It is written in `go`. **It is currently under development, and may change in future versions.**

# Why glurmo?

In short, I have had to run quite a few simulations over the course of my PhD, and found that writing a simulation manager sped up the process quite a bit. But if you'd like a more detailed answer, see below.

## Organization

Running a slurm job can produce quite a bit of output- an `.out` file from slurm itself, an `.Rout` file if you're running an R script, a different file for errors (if you so choose), and, of course, the actual artifact of your simulation. `glurmo` establishes a pre-determined structure for a simulation study, so all of these files are well organized and easy to find.

## Reusability

This standardized structure is one example of the primary goal of `glurmo`: more re-usability. Having standardized simulation directory structures makes it easy to write code to summarize simulation results that will work on any simulation with the same type of result. More importantly, `glurmo` makes it easy to re-use and modify script templates and slurm templates with different settings by tweaking a few values.

## DRY

Another primary goal of `glurm` is to make your simulations [DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself)-er. By defining the script template, slurm template, and settings in three files that can produce an arbitrary number of simulations, changing a single value can change a setting across an entire simulation (or even set of simulations).

# Why not job arrays?

There's nothing that `glurmo` can do that job arrays can't. But having used both, there are a few reasons why I still prefer `glurmo`. 

## Readability

`glurmo` uses a templating engine, so the resulting script and slurm templates are much more readable than if you rely on the index of a job array. 

## Ease of use

I am almost certainly biased here, but I find `glurmo` easier to use than job arrays, particularly when it comes to cancelling jobs. With `glurmo`, you can run a single command and cancel a certain number of jobs in specific states, without having to check job IDs. All you need to know is the base directory. You also don't need to keep track of which simulations you have run in an array or not.

## Power (eventually)

`glurmo` uses templates to define `slurm` submission scripts. This means that in theory you can even set the resources requested by the simulation according to the settings themselves. However, I still need to decide on the interface for this. 

# Why is it *called* glurmo?

`slurm` is a reference to [Futurama](https://en.wikipedia.org/wiki/Slurm_Workload_Manager#History), so I thought it was only fitting to make the name of glurmo a Futurama reference from the same episode.
