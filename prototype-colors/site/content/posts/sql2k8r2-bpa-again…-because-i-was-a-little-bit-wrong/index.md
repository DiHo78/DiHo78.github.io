---
title: "SQL2K8R2 BPA again…. because I was a little bit wrong"
date: "2010-09-18T13:27:18+00:00"
url: "/sql2k8r2-bpa-again…-because-i-was-a-little-bit-wrong/"
lang: "en"
categories: ["SQL Server"]
tags: ["SQL2k8R2 BPA"]
---
I have to apologize, because I made a little mistake in my previous post.

It is correct that the MBCA has a problem running properly when group policies are defined, BUT I was wrong with the remote scanning.

Let me explain my little “mistake”:

If you´re running the MBCA there are two ways scanning a machine.

1\. choosing ‘Connect to another Computer’ as shown here

[![BPA_1](bpa_15b45d.jpg)](bpa_15b45d.jpg)

Then you´ll probably get the error as described in the previous post (the execution policy stuff)

So, the 2nd way to go is:

Choose the SQL Server 2008 BPA from Dropdown menu and click ‘Start Scan’

[![BPA_2](bpa_25b45d1.jpg)](bpa_25b45d1.jpg)

In the next menu, you can define parameters for scanning as well as another computer!

[![BPA_3](bpa_35b45d1.jpg)](bpa_35b45d1.jpg)

So, sorry that I was wrong saying you cannot scan remote computers easily.

I will try to play around a little bit with the BPA. If I find more interesting stuff or issues, I will let you know.
