const imageForm = document.querySelector("#imageForm")
const imageInput = document.querySelector("#imageInput")

imageForm.addEventListener("submit", async e => {
    e.preventDefault()
    const file = imageInput.files[0]

    const url = "https://coolcar.s3.us-west-2.amazonaws.com/62aec076c8f95a7b7da75ef0/62aee17a93de6700ed66771a?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIAV3NUNFUB3ISTDRSK%2F20220619%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20220619T084234Z&X-Amz-Expires=500&X-Amz-SignedHeaders=host&x-id=PutObject&X-Amz-Signature=3e37579056e165091256855d99c0fabdbdd7ec2b2931e756401fd22e6dd954a3"


    await fetch(url, {
        method: "PUT",
        body: file
    })
    const imageUrl = url.split("?")[0]
    console.log("imageUrl", imageUrl)
})

