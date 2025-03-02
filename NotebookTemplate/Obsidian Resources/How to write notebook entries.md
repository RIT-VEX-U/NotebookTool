# Introduction
This document is meant as an introduction to the format of notebook entries. In order to be automatically collected and rendered, the entries must follow the format outlined in these steps.

The steps roughly follow the engineering design process that VEX Judges are told to evaluate notebooks on. 

![[Pasted image 20240606191851.png|center|300]]
*The rubric on which the engineering process is evaluated*


For the sake of demonstration, I will be using a puzzle based project
# Identify the Problem, Issue, or Subproject

This is the first step in the design process. Before you design anything, you must have a problem. These can be of different scales - something as small as one structural member is cracking so we need to develop a replacement to something larger like an entire subsystem of a robot. For the sake of work distribution, documentation, and collaboration, lean towards many small-medium sized projects rather than a large one like Build a Robot.
For example:
- 
## Setup

For this example, I have the following problem.
I can't find the puzzle piece that finishes this flower.
![[Pasted image 20240829172700.png|center|300]]

As this is a hardware project, I will make a folder for it under Hardware Entries. 

![[Pasted image 20240829172819.png|center]]

In this folder, I will make a note describing the problem and providing additional context for why it is a problem. 

## Create a Note

To create this note, right click on the folder that was just created and select new note. Give this note a title that describes the project. A standard title format we have settled on is `Month-Day-Year Title Words`
![[Pasted image 20240829221142.png]]
This keeps important information available at a glance. 

You should see a frightengly empty canvas in front of you on which to paint your masterpiece. 
The first step when creating a new notebook entry is to select a template for the type of note this is. Templates can be added by accessing the command pallete (Ctrl-P) and typing Insert Template

![[Pasted image 20240829173816.png]]

As this is an "identify the problem" entry for the hardware notebook, select the `hardware identify the problem` template. If you are writing a software entry, use the `software identify the problem` template
![[Pasted image 20240829173904.png]]
## Fill in the metadata
This will bring in a helpful block that you should fill in with information.
![[Pasted image 20240829174451.png|center|500]]
- `notebook` is already filled in as Hardware and the `process-step` has also been filled in from the template. Process step is the step of the engineering design process this entry represents. 
- You can enter your name and your co-authors names in the `authors` field
- Select the `date` that the entry was first written on. 
- Select finished when you are done writing and are ready to have the entry reviewed
- Proofread should be left empty until at least one other person has looked over this entry
- Icon and IconColor are just for controlling how the note appears in the explorer pane - they can be left alone

## Fill in the entry
Start of an identify the problem with a short summary of the problem. A sentence or two is fine here. 
![[Pasted image 20240829174926.png|center|500]]
Underneath this short description, delve into detail about what the problem entails, and why its a problem. Feel free to use Markdown headers to logically break up sections of the entry. Add images, tables, and whatever you need to tell the story of why this is a problem that the team needs to solve.

![[Pasted image 20240829175423.png|center|500]]

If this all went smoothly, congrats! You now have a problem on your hands 🎉

# Brainstorm, Diagram, and Prototype

Now that you have a problem, you have to start working on a solution. Time to start thinking, drawing and prototyping. There will often be more than one of these note as work occurs over multiple meetings so you should split up notes by date. 

## Creating the note
The folder for the project is already made so we just have to make a note for this step. Again, right clicking on the folder and pressing new note will create a note for the project. 
![[Pasted image 20240829221205.png|center]]
Just like the identifying the problem note, we have templates for brainstorming. 

![[Pasted image 20240829221334.png]]
For these notes, write about what people think up as solutions for the problem. Include images and tables to better describe anything visual. And please, do a better job than I did for this example entry.

![[Pasted image 20240829221938.png]]

Sometimes, the design process is not linear. Don't be afraid to brainstorm big ideas more than once just make sure to write it all down. 

Additionally, if the brainstorming is substantial, feel free to create different entries for each idea to flesh them out in more detail and to avoid too much happening in one note. The note booking entries are a good example of this.

# Select Best Solution

Once you've brainstormed and prototyped possible solutions, its time to get down to business. Often only one idea will be chosen to pursue due to finite time, people, and resources, but sometimes the team can decide to pick multiple to pursue. 

This type of entry describes the reasons why any solution would be preferred against any other. Often, a meeting will happen when all the members of the project discuss pros and cons of any possible solution and maybe even make a ✨ decision matrix✨. Use these entries to document and standardize that process. This is important not just to keep your discussions grounded in reality but is also vitally important when looking back for past issues and successes.

The process for making one of these entries is similar to previous notes. 
1. Create a new note in the project folder
2. insert the `hardware select best` template
3. fill in the date and author fields
4. document which of possible ideas and prototypes the team chooses as well as the arguments and discussion about the decision. 
5. check the finished box when the note is complete
6. have people proofread it and enter their name in the proofread-by section of the header

![[Pasted image 20240829223557.png|center|600]]

Now is a good time to mention the table at the top of the notes. We have a certain [[RIT VEX U Notebook Standards|format]] we want all notebook entries to abide by. This table will show any violations that we are able to check automatically. It will hide itself in preview mode and in the final notebook so there is never any need to delete the codeblock that spawns it.
![[Pasted image 20240829223745.png|center]]


# Build and Program

Now that we have a path to follow, we better get following. For the sake of being generic between hardware and software, we have chosen to call this step a "project update". There will often be many updates for any given project so make as many of these entries as days you spend working. Because of this, these entries can often be short. However, try to include at least some images to give some ideas as to progress along side text describing the status of the project. 

1. Create a new note in the project folder
2. insert the `hardware update` or `software update` template
3. fill in the date and author fields
4. document the work that was done during this meeting. 
5. check the finished box when the note is complete
6. have people proofread it and enter their name in the proofread-by section of the header

# Test

Testing is an important part of the design process. Just like build updates, there will be many tests along the way as the design is refined and built. A test entry can be short and sweet for a check if a piece fits on the robot or something more involved like testing a component to destruction and data for each load that was put upon it.

1. Create a new note in the project folder
2. insert the `hardware test` or `software test` template
3. fill in the date and author fields
4. document the test that was performed and its results.
5. If needed, document possible ideas that come to mind in response to the test. 
6. check the finished box when the note is complete
7. have people proofread it and enter their name in the proofread-by section of the header

# Project Overview

A project overview is like a mini preview of the notebook specific to a project. It will put the entries from a project in chronological order in the format of the actual notebook. 

To create a project summary note, right click on the folder to add a new note. Then insert the `Project Overview` template. The project overview can not be edited directly - rather it will update to reflect the notes in its project directory. 

# Conclusion

![[Pasted image 20240829235045.png|center]]
Congrats 🎉. You have done a lot of work on this project. I hope it turned out well and that it contains a lot more detail than the [[8-29-24 Missing Puzzle Piece|Sample Project]]
