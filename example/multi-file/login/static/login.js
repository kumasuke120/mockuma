(function() {
	"use strict";

	if (typeof(Storage) === "undefined") { // checks for web storage
		$(".form-login").text("Oops, no HTML5 Web Storage support.");
		return;
	}

	const USERNAME_STORE_KEY = "mockuma.example.login.username";
	const PASSWORD_STORE_KEY = "mockuma.example.login.password";
	function loadRememberedData() {
		let username = localStorage.getItem(USERNAME_STORE_KEY);
		if (username != null) {
			$("#inputUsername").val(username);
			$("#checkboxRememberMe").prop("checked", true);
		}
		let password = localStorage.getItem(PASSWORD_STORE_KEY);
		if (password != null) {
			$("#inputPassword").val(password);
		}
	}
	loadRememberedData();

	function uuidv4() {
		return ([1e7]+-1e3+-4e3+-8e3+-1e11).replace(/[018]/g, 
			c => (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16)
			);
	}

	// toast helper
	const toastDst = $(".toasts > div");
	function Toast(title, message) {
		this.eleId = uuidv4();
		this.html = `<div id="${this.eleId}" class="toast" role="alert" aria-live="assertive" aria-atomic="true">
		<div class="toast-header">
		<img src="/login/favicons/favicon-16x16.png" class="rounded mr-2" alt="favicon">
		<strong class="mr-auto">${title}</strong>
		<small class="time-elapsed text-muted">just now</small>
		<button type="button" class="ml-2 mb-1 close" data-dismiss="toast" aria-label="Close">
		<span aria-hidden="true">&times;</span>
		</button>
		</div>
		<div class="toast-body">
		${message}
		</div>
		</div>`
	}
	Toast.prototype.setDelay = function(delay){
		this.delay = delay
	};
	Toast.prototype.show = function() {
		toastDst.append(this.html);
		let ele = $("#"+this.eleId);

		let seconds = 0;
		let pid = setInterval(function() {
			ele.find(".time-elapsed").text(`${++seconds} second${seconds > 1 ? "s" : ""} ago`);
		}, 1000);

		ele.toast({ "delay": this.delay || 3000 });
		ele.toast("show");
		ele.on("hidden.bs.toast", function() {
			$(this).remove();
			clearInterval(pid);
		});
	};

	// login form
	function onLoginClicked(e) {
		let thisBtn = $(this);

		// checks form
		let form = thisBtn.parents("form")[0];
		if (!form.checkValidity()) return;

		e.preventDefault();

		// locks all form controls
		$(form).find(":input").prop("disabled", true);

		// performs login
		let username = $("#inputUsername").val();
		let password = $("#inputPassword").val();
		let isRemember = $("#checkboxRememberMe").is(":checked");

		if (!isRemember) {
			localStorage.clear();
		}

		$.ajax({
			url: "/api/login",
			type: "post",
			data: { username },
			headers: {
				"Authorization": `Basic ${btoa(username + "/" + password)}`
			},
			success: function(data) {
				if (isRemember) {
					localStorage.setItem(USERNAME_STORE_KEY, username);
					localStorage.setItem(PASSWORD_STORE_KEY, password);
				}

				let t = new Toast("Welcome!", data.message);
				t.setDelay(5000);
				t.show();

				thisBtn.text("Redirecting...");
				setTimeout(function() {
					document.location.href = "/";
				}, 3000);
			},
			error: function(data) {
				$(form).find(":input:disabled").prop("disabled", false);

				let resp = data.responseJSON;
				new Toast("Error " + resp.code, resp.message).show();
			}
		});
	}
	$(".form-login button[type='submit']").click(onLoginClicked);
})();