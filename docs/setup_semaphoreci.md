<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Setup SemaphoreCI

We use semaphore ci to build artifacts and run tests.

1. Navigate to http://semaphoreci.com and create a new account and organization

<img width="1555" alt="Screenshot 2023-08-05 at 9 05 29 AM" src="https://github.com/networknext/next/assets/696656/53569f54-6625-4cea-8aec-d9b5f8755b3d">

You want to create a new account with the "startup" plan and it should be linked to your github account and organization.

2. Create a new project called 'next'

<img width="1046" alt="Screenshot 2023-08-05 at 9 08 00 AM" src="https://github.com/networknext/next/assets/696656/5de1210b-3dec-48d3-80bc-d9115c1b087b">

3. Link the project to your forked 'next' repository.

<img width="1040" alt="Screenshot 2023-08-05 at 9 12 06 AM" src="https://github.com/networknext/next/assets/696656/1bae25f5-162e-41cd-9b9b-a46c465b5fa4">

4. Configure the project

The next repository already contains semaphoreci configuration in `.semaphore/semaphore.yml`

<img width="865" alt="Screenshot 2023-08-05 at 9 13 43 AM" src="https://github.com/networknext/next/assets/696656/e97c95b0-9b37-4d47-bf7c-ce4d8700ca5f">

Select "I will use the existing configuration".

5. Trigger a build and verify that it succeeds

Make asy change in your forked repository and push it to origin.

Semaphore will automatically trigger a build:

<img width="941" alt="Screenshot 2023-08-05 at 9 18 43 AM" src="https://github.com/networknext/next/assets/696656/1df2b187-cfe3-42f3-a91a-c3b06032af02">

5. Drill down into the build in progress
   
Click on the build and you should see the build in progress:

<img width="899" alt="Screenshot 2023-08-05 at 9 18 54 AM" src="https://github.com/networknext/next/assets/696656/8a71dafb-3d34-4196-b3e9-97fb8fc7797f">

5. Verify that the build succeeds

All build items should complete successfully and turn green in around one minute.

<img width="679" alt="Screenshot 2023-08-05 at 9 20 08 AM" src="https://github.com/networknext/next/assets/696656/7eddd25f-405f-4524-961b-838bec534d8a">

6. Build and run extended tests

...

_todo: you are now ready to go to the next step_
