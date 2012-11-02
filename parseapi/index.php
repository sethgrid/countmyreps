<?php

if ($_POST){
    include ("post_data.php");
}
else if ($_GET['record']){
    include ("get_data.php");
}
else{
    header('Location: http://www.countmyreps.com/');
}
