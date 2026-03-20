document.querySelector<HTMLInputElement>("input[name='q']")
  ?.addEventListener("input", function () {
    const val = this.value;
    const matched = Array.from(
      document.querySelectorAll<HTMLOptionElement>("#searched-words option"),
    ).some((opt) => opt.value.toUpperCase() === val.toUpperCase());
    if (matched) {
      alert(this.value);
    }
  });
