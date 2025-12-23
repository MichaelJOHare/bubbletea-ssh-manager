# Step 1: Go to msys2.org and download the installer

<img width="302" height="94" alt="Image" src="https://github.com/user-attachments/assets/b557634a-5b74-4d58-b6e2-b69baac26c60" />

This changes your "Home" to be in C:\msys64\home\username\ when using a bash shell, keep that in mind. You should have the .bin folder that contains the .exe for this project in that "Home" not the Windows home. Also be sure that the .bashrc, .bash_profile, and .profile files are there too.

__IMPORTANT__: If you already have an SSH config make sure to copy the .ssh folder from your Windows home to the MSYS2 home otherwise your saved hosts etc. won't show up and you'll have to add them manually.

🟩 Correct: C:\msys64\home\username\\.bin\vmsmenu.exe  <--- MSYS2 Home

🟥 Incorrect: C:\Users\username\\.bin\vmsmenu.exe  <--- Windows Home

<br>
<br>

# Step 2: Use pacman package manager to update packages (Do this twice!)

<img width="302" height="58" alt="Image" src="https://github.com/user-attachments/assets/63f6da62-f694-43a5-8be2-32e72c26de80" />

    ...

<img width="302" height="58" alt="Image" src="https://github.com/user-attachments/assets/63f6da62-f694-43a5-8be2-32e72c26de80" />

<br>
<br>

# Step 3: Use pacman to install OpenSSH and telnet (inetutils)

<img width="302" height="54" alt="Image" src="https://github.com/user-attachments/assets/83b51732-4784-4dbd-a752-232766b4c5b2" />
<br>
<img width="302" height="55" alt="Image" src="https://github.com/user-attachments/assets/619d3ff1-b5da-4296-8bee-a0cc1756d19a" />
<br>
<br>
Update packages again, it probably won't find anything to update which is fine, but it's just to be sure.

<img width="302" height="58" alt="Image" src="https://github.com/user-attachments/assets/63f6da62-f694-43a5-8be2-32e72c26de80" />

<br>
<br>

# Step 4: Download Windows Terminal from the Microsoft Store

<img width="707" height="293" alt="Image" src="https://github.com/user-attachments/assets/e7f84c1d-e38b-44ea-b10b-85e269ff8a34" />

<br>
<br>

# Step 5: Copy .bash_profile, .bashrc, and .profile files
    
This is entirely optional, they just have some changes like allowing you to type vmsmenu at the command prompt instead of having to launch the .exe directly. 
If you want, you can download Go and build/install this yourself and not worry about it (although you might run into $PATH issues).

<br>
<br>

# Step 6: Replace Windows Terminal settings.json with the one in this repo

This is optional, it includes keybinding changes so that it can mimic "application keypad mode" where the keypad will send escape sequences  
instead of the actual number on the keypad. This is very useful for editors on OpenVMS. If you do skip this part, just make sure the CommandLine 
for the Windows Terminal profile you make to launch the MSYS2 bash shell looks something like this:
<br>
<br>
__C:\msys64\msys2_shell.cmd -defterm -here -no-start -ucrt64__

<br>
<br>

# 🎉 Finished! 🎉

You should be able to type vmsmenu at the shell prompt to use the session manager to connect to your saved hosts. 
You can also type addhost to use an interactive menu to save sessions into the config files so that vmsmenu can use them.

<img width="707" height="309" alt="Image" src="https://github.com/user-attachments/assets/c9653f11-ad14-4a8c-b381-2458a0c39b8c" />
