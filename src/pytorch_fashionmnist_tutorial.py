# -*- coding: utf-8 -*-
"""Tutorial PyTorch: phân loại FashionMNIST bằng mạng MLP.

Cài đặt:
    python -m pip install torch torchvision

Chạy:
    python pytorch_fashionmnist_tutorial.py

Lần chạy đầu tiên sẽ tải dữ liệu FashionMNIST (~30 MB) từ Internet.
"""

import torch
from torch import nn
from torch.utils.data import DataLoader, random_split
from torchvision import datasets
from torchvision.transforms import v2

BATCH_SIZE = 64
EPOCHS = 5
LEARNING_RATE = 1e-3
SEED = 42
MODEL_PATH = "fashion_mlp_best.pth"

CLASS_NAMES = [
    "T-shirt/top", "Trouser", "Pullover", "Dress", "Coat",
    "Sandal", "Shirt", "Sneaker", "Bag", "Ankle boot",
]


def get_device():
    if torch.cuda.is_available():
        return torch.device("cuda")
    if hasattr(torch.backends, "mps") and torch.backends.mps.is_available():
        return torch.device("mps")
    return torch.device("cpu")


def make_dataloaders():
    transform = v2.Compose([
        v2.ToImage(),
        v2.ToDtype(torch.float32, scale=True),
    ])

    full_train = datasets.FashionMNIST(
        root="data", train=True, download=True, transform=transform
    )
    test_set = datasets.FashionMNIST(
        root="data", train=False, download=True, transform=transform
    )

    train_set, val_set = random_split(
        full_train, [54_000, 6_000],
        generator=torch.Generator().manual_seed(SEED),
    )

    train_loader = DataLoader(train_set, batch_size=BATCH_SIZE, shuffle=True)
    val_loader = DataLoader(val_set, batch_size=BATCH_SIZE, shuffle=False)
    test_loader = DataLoader(test_set, batch_size=BATCH_SIZE, shuffle=False)
    return train_loader, val_loader, test_loader


class FashionMLP(nn.Module):
    def __init__(self):
        super().__init__()
        self.network = nn.Sequential(
            nn.Flatten(),
            nn.Linear(28 * 28, 256),
            nn.ReLU(),
            nn.Linear(256, 128),
            nn.ReLU(),
            nn.Linear(128, 10),
        )

    def forward(self, images):
        # Output là logits (điểm số thô), chưa áp dụng softmax.
        return self.network(images)


def train_one_epoch(model, loader, loss_fn, optimizer, device):
    model.train()
    total_loss = 0.0
    total_correct = 0
    total_samples = 0

    for images, labels in loader:
        images = images.to(device)
        labels = labels.to(device)

        optimizer.zero_grad(set_to_none=True)
        logits = model(images)
        loss = loss_fn(logits, labels)
        loss.backward()
        optimizer.step()

        n = labels.size(0)
        total_samples += n
        total_loss += loss.item() * n
        total_correct += (logits.argmax(dim=1) == labels).sum().item()

    return total_loss / total_samples, total_correct / total_samples


@torch.inference_mode()
def evaluate(model, loader, loss_fn, device):
    model.eval()
    total_loss = 0.0
    total_correct = 0
    total_samples = 0

    for images, labels in loader:
        images = images.to(device)
        labels = labels.to(device)

        logits = model(images)
        loss = loss_fn(logits, labels)

        n = labels.size(0)
        total_samples += n
        total_loss += loss.item() * n
        total_correct += (logits.argmax(dim=1) == labels).sum().item()

    return total_loss / total_samples, total_correct / total_samples


def main():
    torch.manual_seed(SEED)
    device = get_device()
    print("Thiết bị:", device)
    print("PyTorch:", torch.__version__)

    train_loader, val_loader, test_loader = make_dataloaders()
    first_images, first_labels = next(iter(train_loader))
    print("Shape ảnh:", tuple(first_images.shape))  # (64, 1, 28, 28)
    print("Shape nhãn:", tuple(first_labels.shape))  # (64,)

    model = FashionMLP().to(device)
    loss_fn = nn.CrossEntropyLoss()
    optimizer = torch.optim.Adam(model.parameters(), lr=LEARNING_RATE)

    best_val_acc = -1.0
    for epoch in range(1, EPOCHS + 1):
        train_loss, train_acc = train_one_epoch(
            model, train_loader, loss_fn, optimizer, device
        )
        val_loss, val_acc = evaluate(model, val_loader, loss_fn, device)

        print(
            f"Epoch {epoch:02d}/{EPOCHS} | "
            f"train loss: {train_loss:.4f}, train acc: {train_acc:.2%} | "
            f"val loss: {val_loss:.4f}, val acc: {val_acc:.2%}"
        )

        if val_acc > best_val_acc:
            best_val_acc = val_acc
            torch.save(model.state_dict(), MODEL_PATH)
            print("  -> Đã lưu mô hình tốt nhất:", MODEL_PATH)

    # Chỉ đánh giá tập test sau khi chọn mô hình bằng validation.
    model.load_state_dict(
        torch.load(MODEL_PATH, map_location=device, weights_only=True)
    )
    test_loss, test_acc = evaluate(model, test_loader, loss_fn, device)
    print(f"Test loss: {test_loss:.4f} | Test accuracy: {test_acc:.2%}")

    # Dự đoán một ảnh chưa sử dụng để huấn luyện.
    image, true_label = test_loader.dataset[0]
    model.eval()
    with torch.inference_mode():
        logits = model(image.unsqueeze(0).to(device))
        probabilities = torch.softmax(logits, dim=1)
        pred_index = probabilities.argmax(dim=1).item()
        confidence = probabilities[0, pred_index].item()

    print("Nhãn thật:", CLASS_NAMES[true_label])
    print("Dự đoán:", CLASS_NAMES[pred_index])
    print(f"Softmax score: {confidence:.2%}")


if __name__ == "__main__":
    main()
