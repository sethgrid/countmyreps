<?php
include("../includes/DataStore.php");

$DataStore = new DataStore;
$content = '';

$user = $_POST['email'];

if ($DataStore->user_exists($user)){
    $content = "User Found";
}
else{
    $content = "No User Found";
}


// look up user
// if found, display reps

// else show error

?>
<html>
<head>
    <title>CountMyReps</title>
    <style  type="text/css">
        body{
            background-color: #E8E8E8;
        }
        div.center{
            margin: auto;
            background-color: white;
            width: 800px; 
            border: 1px solid #C8C8C8; 
            padding-top: 10px;
            padding-bottom: 20px;
       }
       div.inner{
            margin: auto;
            text-align: center;
            padding-top: 300px;
            color: #666362;
       }
    </style>
</head>
<body>

<div class="center">
    <div class="inner">
    <?php
        echo $content;
    ?>
    </div>
</div>

</body>
</html>
