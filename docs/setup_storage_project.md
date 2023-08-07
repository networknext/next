<img src="https://static.wixstatic.com/media/799fd4_0512b6edaeea4017a35613b4c0e9fc0b~mv2.jpg/v1/fill/w_1200,h_140,al_c,q_80,usm_0.66_1.00_0.01/networknext_logo_colour_black_RGB_tightc.jpg" alt="Network Next" width="600"/>

<br>

# Setup storage project

In this section we will setup a google cloud project called "Storage" where we will store build artifacts. We will upload these files to cloud storage using semaphoreci jobs.

1. Click on the project selector at the top left of the screen at http://console.cloud.google.com:

<img width="919" alt="Screenshot 2023-08-06 at 9 05 32 PM" src="https://github.com/networknext/next/assets/696656/90fa1962-a03b-4192-a823-fa955853496f">

2. Select "NEW PROJECT"
   
<img width="751" alt="Screenshot 2023-08-06 at 9 09 37 PM" src="https://github.com/networknext/next/assets/696656/2443b842-73ad-4dd1-b6d6-cd01dd757619">

3. Create a new project called "Storage"

<img width="553" alt="image" src="https://github.com/networknext/next/assets/696656/1f57d358-919b-4f80-87c5-1d3d99ac548c">

Once the project is created it is assigned a project id:

<img width="493" alt="Screenshot 2023-08-06 at 9 12 31 PM" src="https://github.com/networknext/next/assets/696656/5872da7d-df9e-443f-afac-af61e4bc1c84">

For example: storage-395201

Save this project id somewhere for later.

4. Click on "Cloud Storage" -> "Buckets" in the google cloud nav menu
   
<img width="530" alt="Screenshot 2023-08-06 at 9 17 30 PM" src="https://github.com/networknext/next/assets/696656/11eb1b6c-7f59-4e32-8e55-d06d7e21f736">

5. Create a cloud storage bucket for development artifacts
   
<img width="526" alt="Screenshot 2023-08-06 at 9 19 34 PM" src="https://github.com/networknext/next/assets/696656/18b7a2fb-ed49-48b9-b781-cd9dc8ea6d18">

