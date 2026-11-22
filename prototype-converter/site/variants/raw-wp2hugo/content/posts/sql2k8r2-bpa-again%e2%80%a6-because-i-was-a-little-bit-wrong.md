---
_edit_last: "13892234"
_oembed_5b7bfeba10fc1be660f36fa8f519d873: '{{unknown}}'
_oembed_7f467e7d17bb293fae1f3aebc230e1ec: '{{unknown}}'
_oembed_34f97ccb558b2568167103932664a156: '{{unknown}}'
_oembed_73aec8056aed68a4923f82673307e237: '{{unknown}}'
_oembed_476747beb97d1b15d37a399b880ba541: '{{unknown}}'
_oembed_a7c3b8135efd80655e79ce1a9a67fed9: '{{unknown}}'
_oembed_b0bb75a0d9e570d90bbaa5a63a78889f: '{{unknown}}'
_oembed_db2f96aafe15a35c466e282e3ba47993: '{{unknown}}'
_wpas_done_ms: "1"
author: feleryan
categories:
  - sql-server
date: "2010-09-18T13:27:18+00:00"
guid: http://feleryan.wordpress.com/2010/09/18/sql2k8r2-bpa-again%e2%80%a6-because-i-was-a-little-bit-wrong
parent_post_id: null
post_id: "14"
tags:
  - sql2k8r2-bpa
title: SQL2K8R2 BPA again&#8230;. because I was a little bit wrong
url: /2010/09/18/sql2k8r2-bpa-again…-because-i-was-a-little-bit-wrong/

---
I have to apologize, because I made a little mistake in my previous post.

It is correct that the MBCA has a problem running properly when group policies are defined, BUT I was wrong with the remote scanning.

Let me explain my little “mistake”:

If you´re running the MBCA there are two ways scanning a machine.

1\. choosing ‘Connect to another Computer’ as shown here

[![BPA_1](/wp-content/uploads/2010/09/bpa_15b45d.jpg?w=300)](/wp-content/uploads/2010/09/bpa_15b45d.jpg)

Then you´ll probably get the error as described in the previous post (the execution policy stuff)

So, the 2nd way to go is:

Choose the SQL Server 2008 BPA from Dropdown menu and click ‘Start Scan’

[![BPA_2](/wp-content/uploads/2010/09/bpa_25b45d1.jpg?w=300)](/wp-content/uploads/2010/09/bpa_25b45d1.jpg)

In the next menu, you can define parameters for scanning as well as another computer!

[![BPA_3](/wp-content/uploads/2010/09/bpa_35b45d1.jpg?w=300)](/wp-content/uploads/2010/09/bpa_35b45d1.jpg)

So, sorry that I was wrong saying you cannot scan remote computers easily.

I will try to play around a little bit with the BPA. If I find more interesting stuff or issues, I will let you know.
